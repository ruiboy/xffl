-- Insert test data into ffl.club table
INSERT INTO ffl.club (name) VALUES
    ('Ruiboys'),
    ('Cheetahs'),
    ('Slashers'),
    ('The Howling Cows')
ON CONFLICT (name) DO NOTHING; 