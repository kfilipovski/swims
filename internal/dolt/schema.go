package dolt

const schemaSQL = `
CREATE TABLE IF NOT EXISTS swimmers (
    swimmer_id   BIGINT PRIMARY KEY,
    full_name    VARCHAR(255) NOT NULL,
    club_name    VARCHAR(255),
    lsc_code     VARCHAR(10),
    age          INT,
    synced_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    times_synced_at   DATE
);

CREATE TABLE IF NOT EXISTS meets (
    id           BIGINT AUTO_INCREMENT PRIMARY KEY,
    meet_name    VARCHAR(255) NOT NULL,
    UNIQUE KEY uq_meet_name (meet_name)
);

CREATE TABLE IF NOT EXISTS events (
    id           BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_code   VARCHAR(20) NOT NULL,
    distance     INT NOT NULL,
    stroke       VARCHAR(10) NOT NULL,
    course       VARCHAR(3) NOT NULL,
    UNIQUE KEY uq_event_code (event_code)
);

CREATE TABLE IF NOT EXISTS times (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    swimmer_id      BIGINT NOT NULL,
    event_id        BIGINT NOT NULL,
    meet_id         BIGINT NOT NULL,
    swim_time       VARCHAR(20) NOT NULL,
    sort_key        VARCHAR(50) NOT NULL,
    age_at_meet     INT,
    power_points    DECIMAL(8,2),
    time_standard   VARCHAR(50),
    lsc_code        VARCHAR(10),
    team_name       VARCHAR(255),
    swim_date       DATE,
    synced_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_time (swimmer_id, event_id, sort_key, swim_date, meet_id),
    INDEX idx_swimmer (swimmer_id),
    INDEX idx_event (event_id),
    INDEX idx_meet (meet_id),
    INDEX idx_date (swim_date)
);
`

func (d *Dolt) EnsureSchema() error {
	return d.SQLBatch(schemaSQL)
}
