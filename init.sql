CREATE TABLE IF NOT EXISTS user_info (
  id UUID PRIMARY KEY,
  telegram_id VARCHAR NOT NULL,
  username VARCHAR,
  firstname VARCHAR,
  lastname VARCHAR,
  CONSTRAINT telegram_unique UNIQUE(telegram_id)
);

CREATE TABLE IF NOT EXISTS audience (
  id UUID PRIMARY KEY,
  number VARCHAR NOT NULL,
  building VARCHAR NOT NULL,
  floor INTEGER NOT NULL,
  suffix VARCHAR,
  CONSTRAINT audience_unique UNIQUE(number, suffix)
);

CREATE TABLE IF NOT EXISTS lesson (
  id UUID PRIMARY KEY,
  name VARCHAR NOT NULL,
  teacher_name VARCHAR,
  kind VARCHAR
);

CREATE TABLE IF NOT EXISTS groups (
  id UUID PRIMARY KEY,
  name VARCHAR NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS schedule (
  id UUID PRIMARY KEY,
  audience_id UUID  NOT NULL REFERENCES audience(id),
  lesson_id UUID  NOT NULL REFERENCES lesson(id),
  week_type VARCHAR NOT NULL,
  week_day VARCHAR NOT NULL,
  lesson_start TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  lesson_end TIMESTAMP WITHOUT TIME ZONE NOT NULL,
  period INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS group_lesson (
  id UUID PRIMARY KEY,
  group_id UUID  NOT NULL REFERENCES groups(id),
  lesson_id UUID  NOT NULL REFERENCES lesson(id),
  CONSTRAINT group_lesson_unique UNIQUE(group_id, lesson_id)
);

CREATE INDEX IF NOT EXISTS schedule_week_type_idx ON schedule USING btree (week_type);
CREATE INDEX IF NOT EXISTS schedule_weekday_idx ON schedule USING btree (week_day);
CREATE INDEX IF NOT EXISTS schedule_period_idx ON schedule USING btree (period);
CREATE INDEX IF NOT EXISTS audience_building_idx ON audience USING btree (building);
CREATE INDEX IF NOT EXISTS audience_floor_idx ON audience USING btree (floor);
