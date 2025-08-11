CREATE TABLE events
(
    id         UUID PRIMARY KEY,
    user_id    UUID,
    date       DATE,
    event_text TEXT
);

CREATE INDEX idx_events_user_id ON events(user_id);
CREATE INDEX idx_events_date ON events(date);

