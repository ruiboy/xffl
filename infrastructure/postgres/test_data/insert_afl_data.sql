-- Insert AFL test data

-- Insert all 18 AFL clubs
INSERT INTO afl.club (name) VALUES
('Adelaide Crows'),
('Brisbane Lions'),
('Carlton Blues'),
('Collingwood Magpies'),
('Essendon Bombers'),
('Fremantle Dockers'),
('Geelong Cats'),
('Gold Coast Suns'),
('Greater Western Sydney Giants'),
('Hawthorn Hawks'),
('Melbourne Demons'),
('North Melbourne Kangaroos'),
('Port Adelaide Power'),
('Richmond Tigers'),
('St Kilda Saints'),
('Sydney Swans'),
('West Coast Eagles'),
('Western Bulldogs')
ON CONFLICT (name) DO NOTHING;