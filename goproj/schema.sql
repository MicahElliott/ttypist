PRAGMA foreign_keys = ON;

BEGIN TRANSACTION;

CREATE TABLE learnable (
  lid	INTEGER	PRIMARY KEY,	-- may be same as "rank" for typing course
  lname	text	NOT NULL,	-- 'charismatic'
  lrank	int	,	-- 10
  defn	text	,	-- 'exercising a compelling charm which inspires devotion in others.'
  diffy	int	,	-- 82  (or "level"?)
  -- lang  text NOT NULL CHECK (lang in ('music', 'english', 'spanish', 'golang', 'zsh')),
  -- grammar and vocab together comprise "language"
  course	text	NOT NULL CHECK (course IN ('typing', 'spelling', 'grammar', 'vocab', 'music', 'math', 'history', 'medicine', 'biology')),
  UNIQUE (lname, course)	ON CONFLICT IGNORE
);
CREATE INDEX idx_learnable_lid ON learnable (lid);

-- INSERT INTO learnable ( lname, lrank, diffy, course, defn ) VALUES ( 'was', 10, 923948, 'typing', NULL );

CREATE TABLE question (
  qid	INTEGER	PRIMARY KEY,
  lid	INTEGER	NOT NULL REFERENCES learnable(lid),
  qtype	text	NOT NULL CHECK (qtype IN ('cloze', 'typed', 'mcq', 'spoken')),
  qtext	text		-- 'The ___ woman captivated her attentive audience.'
  -- distrs	text		-- json: '["person of tall stature", "an automatic geared vehicle"]'
);

CREATE TABLE distractor (
  did	INTEGER	PRIMARY KEY,
  qid	INTEGER	NOT NULL REFERENCES question(qid),
  dtext	text	NOT NULL	--  'person of tall stature'
);

-- Populate questions
-- TODO move into words-populate.sql
-- INSERT INTO question ( qtype, lid ) SELECT 'typed', lid from learnable;
-- INSERT INTO question ( qtype,lid ) VALUES ('typed',10), ('typed',9996);

CREATE TABLE training (
  tstamp	TIMESTAMP	PRIMARY KEY,
  speed	float	NOT NULL,
  accy	float	NOT NULL,
  nqtns	int	NOT NULL,
  style	text		-- new-learnables-only, review-only, speed, accuracy
);

CREATE TABLE encounter (
  eid	INTEGER	PRIMARY KEY,
  -- lid	INTEGER	REFERENCES learnable(lid),
  qid	INTEGER	NOT NULL REFERENCES question(qid),
  tstamp	INTEGER	NOT NULL REFERENCES training(tstamp),	-- foreign key
  estamp	timestamp	NOT NULL,
  entered	text	NOT NULL,
  timer	float	NOT NULL,
  correct	boolean	NOT NULL,
  score	text	,	-- pass, fail, 89, etc
  acty	text
  -- artistid INTEGER REFERENCES artist(artistid),
  -- PRIMARY KEY	(lid, estamp)
);

-- based on number of encounters; entered upon first encounter of learnable
-- recomputed each training
CREATE TABLE strength (
  sid	INTEGER	PRIMARY KEY,
  lid	INTEGER	NOT NULL REFERENCES learnable(lid),
  val	int	NOT NULL	-- 0-100
);

COMMIT;
