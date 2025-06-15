package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/lib/pq"
	"xffl/pkg/events"
)

// PostgresDispatcher implements EventDispatcher using PostgreSQL LISTEN/NOTIFY
// This is completely separate from domain persistence - it only handles events
type PostgresDispatcher struct {
	eventDB     *sql.DB // Separate connection for events only
	subscribers map[string][]events.EventHandler
	listeners   map[string]*pq.Listener
	logger      *log.Logger
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	connStr     string
}

// NewPostgresDispatcher creates a new PostgreSQL-based event dispatcher
// Uses a separate database connection dedicated only to event messaging
func NewPostgresDispatcher(connStr string, logger *log.Logger) (*PostgresDispatcher, error) {
	// Create dedicated connection for events (separate from domain persistence)
	eventDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL for events: %w", err)
	}

	// Configure connection pool for event workload
	eventDB.SetMaxOpenConns(5)  // Small pool for event operations
	eventDB.SetMaxIdleConns(2)
	eventDB.SetConnMaxLifetime(5 * time.Minute)

	return &PostgresDispatcher{
		eventDB:     eventDB,
		subscribers: make(map[string][]events.EventHandler),
		listeners:   make(map[string]*pq.Listener),
		logger:      logger,
		connStr:     connStr,
	}, nil
}

// Start initializes the PostgreSQL event dispatcher
func (d *PostgresDispatcher) Start(ctx context.Context) error {
	d.ctx, d.cancel = context.WithCancel(ctx)

	// Test event database connection
	if err := d.eventDB.PingContext(d.ctx); err != nil {
		return fmt.Errorf("failed to ping event database: %w", err)
	}

	d.logger.Println("PostgresDispatcher started with dedicated event connection")
	return nil
}

// Stop gracefully shuts down the PostgreSQL dispatcher
func (d *PostgresDispatcher) Stop() error {
	if d.cancel != nil {
		d.cancel()
	}

	// Stop all listeners
	d.mu.Lock()
	for _, listener := range d.listeners {
		listener.Close()
	}
	d.mu.Unlock()

	// Wait for all goroutines to finish
	d.wg.Wait()

	// Close event database connection
	var err error
	if d.eventDB != nil {
		err = d.eventDB.Close()
	}

	d.logger.Println("PostgresDispatcher stopped")
	return err
}

// Subscribe registers an event handler for a specific event type
func (d *PostgresDispatcher) Subscribe(eventType string, handler events.EventHandler) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Add handler to subscribers map
	d.subscribers[eventType] = append(d.subscribers[eventType], handler)

	// Create listener for this event type if it's the first subscriber
	if len(d.subscribers[eventType]) == 1 {
		if err := d.createListener(eventType); err != nil {
			return fmt.Errorf("failed to create listener for %s: %w", eventType, err)
		}
	}

	d.logger.Printf("Subscribed handler '%s' to event type '%s'", handler.HandlerName(), eventType)
	return nil
}

// Publish sends an event using PostgreSQL NOTIFY
func (d *PostgresDispatcher) Publish(ctx context.Context, event events.DomainEvent) error {
	// Serialize event data
	eventData, err := json.Marshal(map[string]interface{}{
		"eventType":    event.EventType(),
		"eventVersion": event.EventVersion(),
		"aggregateId":  event.AggregateID(),
		"occurredAt":   event.OccurredAt(),
		"eventData":    event.EventData(),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Convert event type to PostgreSQL channel name (replace dots with underscores)
	channelName := convertEventTypeToChannel(event.EventType())

	// Use NOTIFY to publish event
	// PostgreSQL NOTIFY syntax: NOTIFY channel, 'payload'
	query := "SELECT pg_notify($1, $2)"
	if _, err := d.eventDB.ExecContext(ctx, query, channelName, string(eventData)); err != nil {
		return fmt.Errorf("failed to publish event via NOTIFY: %w", err)
	}

	d.logger.Printf("Published event '%s' to PostgreSQL channel '%s'", event.EventType(), channelName)
	return nil
}

// createListener creates a PostgreSQL listener for the given event type
func (d *PostgresDispatcher) createListener(eventType string) error {
	channelName := convertEventTypeToChannel(eventType)

	// Create listener with error callback
	listener := pq.NewListener(d.connStr, time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
		if err != nil {
			d.logger.Printf("PostgreSQL listener error for %s: %v", channelName, err)
		}
	})

	// Listen to the channel
	if err := listener.Listen(channelName); err != nil {
		listener.Close()
		return fmt.Errorf("failed to listen on channel %s: %w", channelName, err)
	}

	// Store listener for cleanup
	d.listeners[eventType] = listener

	// Start goroutine to handle notifications
	d.wg.Add(1)
	go d.handleNotifications(eventType, listener)

	d.logger.Printf("Created PostgreSQL listener for event type '%s' on channel '%s'", eventType, channelName)
	return nil
}

// handleNotifications processes PostgreSQL notifications for an event type
func (d *PostgresDispatcher) handleNotifications(eventType string, listener *pq.Listener) {
	defer d.wg.Done()
	defer listener.Close()

	d.logger.Printf("Started listening for PostgreSQL notifications on event type '%s'", eventType)

	for {
		select {
		case <-d.ctx.Done():
			return
		case notification := <-listener.Notify:
			if notification != nil {
				d.processNotification(eventType, notification.Extra)
			}
		case <-time.After(90 * time.Second):
			// Ping periodically to keep connection alive
			go func() {
				if err := listener.Ping(); err != nil {
					d.logger.Printf("Failed to ping listener for %s: %v", eventType, err)
				}
			}()
		}
	}
}

// processNotification handles a single PostgreSQL notification
func (d *PostgresDispatcher) processNotification(eventType, payload string) {
	// Parse the event data
	var eventWrapper map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &eventWrapper); err != nil {
		d.logger.Printf("Failed to unmarshal event notification: %v", err)
		return
	}

	// Create generic event from notification
	genericEvent := &GenericEvent{
		eventType:    eventWrapper["eventType"].(string),
		eventVersion: eventWrapper["eventVersion"].(string),
		aggregateId:  eventWrapper["aggregateId"].(string),
		occurredAt:   time.Now(), // Could parse from eventWrapper if needed
		eventData:    eventWrapper["eventData"].(map[string]interface{}),
	}

	// Get handlers for this event type
	d.mu.RLock()
	handlers := d.subscribers[eventType]
	d.mu.RUnlock()

	// Process event with all handlers
	errorCount := 0
	for _, handler := range handlers {
		if err := handler.Handle(d.ctx, genericEvent); err != nil {
			d.logger.Printf("Handler '%s' failed to process event '%s': %v",
				handler.HandlerName(), eventType, err)
			errorCount++
		}
	}

	d.logger.Printf("Processed event '%s' with %d handlers (%d errors)",
		eventType, len(handlers), errorCount)
}

// convertEventTypeToChannel converts event type to valid PostgreSQL identifier
// e.g., "AFL.PlayerMatchUpdated" -> "afl_player_match_updated"
func convertEventTypeToChannel(eventType string) string {
	result := ""
	for _, char := range eventType {
		switch char {
		case '.':
			result += "_"
		case ' ':
			result += "_"
		default:
			if char >= 'A' && char <= 'Z' {
				result += string(char + 32) // Convert to lowercase
			} else {
				result += string(char)
			}
		}
	}
	return result
}

// GenericEvent implements DomainEvent for PostgreSQL-received events
type GenericEvent struct {
	eventType    string
	eventVersion string
	aggregateId  string
	occurredAt   time.Time
	eventData    map[string]interface{}
}

func (e *GenericEvent) EventType() string {
	return e.eventType
}

func (e *GenericEvent) EventVersion() string {
	return e.eventVersion
}

func (e *GenericEvent) AggregateID() string {
	return e.aggregateId
}

func (e *GenericEvent) OccurredAt() time.Time {
	return e.occurredAt
}

func (e *GenericEvent) EventData() map[string]interface{} {
	return e.eventData
}