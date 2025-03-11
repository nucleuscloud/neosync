SET FOREIGN_KEY_CHECKS=0;  -- (Disable FK checks for initial creation to handle ordering)


CREATE TABLE IF NOT EXISTS agency (
    id          INT AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    country     VARCHAR(64),
    founded_year YEAR,
    INDEX (country)
);

CREATE TABLE IF NOT EXISTS astronaut (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    first_name      VARCHAR(50) NOT NULL,
    last_name       VARCHAR(50) NOT NULL,
    birth_date      DATE NOT NULL,
    nationality     VARCHAR(50),
    status          ENUM('Active','Retired','Deceased') DEFAULT 'Active',
    agency_id       INT,
    first_mission_id INT,  -- FK to mission (circular dependency handled later)
    CONSTRAINT uq_astronaut_name UNIQUE (first_name, last_name, birth_date),
    CONSTRAINT chk_min_birth_date CHECK (birth_date >= '1900-01-01'),
    FOREIGN KEY (agency_id) REFERENCES agency(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS spacecraft (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    type            ENUM('Orbiter','Lander','Rover','Crewed','Probe','Station','Other') NOT NULL,
    capacity        INT,  -- number of crew it can carry (NULL or 0 if unmanned)
    status          ENUM('Operational','In Mission','Retired','Lost') DEFAULT 'Operational',
    agency_id       INT,
    last_mission_id INT,  -- FK to mission (circular dependency handled later)
    CONSTRAINT uq_spacecraft_name UNIQUE (name),
    FOREIGN KEY (agency_id) REFERENCES agency(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS celestial_body (
    id           INT AUTO_INCREMENT PRIMARY KEY,
    name         VARCHAR(100) NOT NULL,
    body_type    ENUM('Star','Planet','Dwarf Planet','Moon','Asteroid','Comet') NOT NULL,
    mass         DOUBLE,  -- mass could be in kg
    radius       DOUBLE,  -- radius in km
    parent_body_id INT NULL,
    CONSTRAINT uq_body_name UNIQUE (name),
    FOREIGN KEY (parent_body_id) REFERENCES celestial_body(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    CONSTRAINT chk_positive_mass CHECK (mass >= 0 OR mass IS NULL));

CREATE TABLE IF NOT EXISTS launch_site (
    id          INT AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    location    VARCHAR(255),  -- e.g., "Cape Canaveral, USA"
    location_coord POINT NOT NULL,      -- spatial coordinate (latitude/longitude)
    country     VARCHAR(50),
    CONSTRAINT uq_site_name UNIQUE (name),
    SPATIAL INDEX (location_coord)
);

CREATE TABLE IF NOT EXISTS mission (
    id               INT AUTO_INCREMENT PRIMARY KEY,
    name             VARCHAR(100) NOT NULL,
    mission_code     VARCHAR(50),
    mission_type     ENUM('Manned','Unmanned') NOT NULL,
    status           ENUM('Planned','Active','Completed','Aborted','Failed') DEFAULT 'Planned',
    launch_date      DATE NOT NULL,
    return_date      DATE,
    spacecraft_id    INT NOT NULL,
    destination_id   INT NOT NULL,
    launch_site_id   INT NOT NULL,
    primary_agency_id INT,
    commander_id     INT,  -- FK to astronaut
    CONSTRAINT uq_mission_code UNIQUE (mission_code),
    CONSTRAINT chk_mission_dates CHECK (return_date IS NULL OR return_date >= launch_date),

    FOREIGN KEY (spacecraft_id) REFERENCES spacecraft(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    FOREIGN KEY (destination_id) REFERENCES celestial_body(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    FOREIGN KEY (launch_site_id) REFERENCES launch_site(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    FOREIGN KEY (primary_agency_id) REFERENCES agency(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    FOREIGN KEY (commander_id) REFERENCES astronaut(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL
);




CREATE TABLE IF NOT EXISTS mission_crew (
    mission_id   INT NOT NULL,
    astronaut_id INT NOT NULL,
    role         ENUM('Commander','Pilot','Engineer','Scientist','Specialist') NOT NULL,
    PRIMARY KEY (mission_id, astronaut_id),
    FOREIGN KEY (mission_id) REFERENCES mission(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    FOREIGN KEY (astronaut_id) REFERENCES astronaut(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    INDEX idx_mc_astronaut (astronaut_id)  -- index for queries by astronaut
);

CREATE TABLE IF NOT EXISTS research_project (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    title           VARCHAR(200) NOT NULL,
    description     TEXT,
    start_date      DATE,
    end_date        DATE,
    lead_astronaut_id INT,
    FOREIGN KEY (lead_astronaut_id) REFERENCES astronaut(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL,
    CONSTRAINT uq_project_title UNIQUE (title),
    FULLTEXT INDEX idx_project_desc (title, description)
);

CREATE TABLE IF NOT EXISTS project_mission (
    project_id  INT NOT NULL,
    mission_id  INT NOT NULL,
    PRIMARY KEY (project_id, mission_id),
    FOREIGN KEY (project_id) REFERENCES research_project(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    FOREIGN KEY (mission_id) REFERENCES mission(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS mission_log (
    log_id      INT AUTO_INCREMENT PRIMARY KEY,
    mission_id  INT NOT NULL,
    log_time    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    event       VARCHAR(255) NOT NULL,
    FOREIGN KEY (mission_id) REFERENCES mission(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    INDEX idx_log_mission (mission_id)
);


-- 2. Add Foreign Keys for circular references (after both tables exist)

ALTER TABLE astronaut
    ADD FOREIGN KEY (first_mission_id) REFERENCES mission(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL;

ALTER TABLE spacecraft
    ADD FOREIGN KEY (last_mission_id) REFERENCES mission(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL;




-- 1) OBSERVATORY
CREATE TABLE IF NOT EXISTS observatory (
    id                INT AUTO_INCREMENT PRIMARY KEY,
    name              VARCHAR(100) NOT NULL,
    agency_id         INT NOT NULL, 
    launch_site_id    INT NULL,     
    location_coord    POINT NULL,   
    status            ENUM('Active','Under Maintenance','Retired') NOT NULL DEFAULT 'Active',
    CONSTRAINT uq_observatory_name UNIQUE (name)
);

-- 2) TELESCOPE
CREATE TABLE IF NOT EXISTS telescope (
    id                INT AUTO_INCREMENT PRIMARY KEY,
    observatory_id    INT NOT NULL,
    name              VARCHAR(100) NOT NULL,
    telescope_type    ENUM('Optical','Radio','Infrared','UV','X-Ray','Other') NOT NULL DEFAULT 'Optical',
    mirror_diameter_m DOUBLE,
    status            ENUM('Operational','Damaged','Retired') DEFAULT 'Operational',
    CONSTRAINT uq_telescope_name UNIQUE (name)
);

-- 3) INSTRUMENT
CREATE TABLE IF NOT EXISTS instrument (
    id                 INT AUTO_INCREMENT PRIMARY KEY,
    name               VARCHAR(100) NOT NULL,
    instrument_type    ENUM('Camera','Spectrometer','Sensor','Module','Other') NOT NULL,
    telescope_id       INT NULL, 
    spacecraft_id      INT NULL,
    status            ENUM('Available','In Use','Damaged','Retired') DEFAULT 'Available',
    CONSTRAINT uq_instrument_name UNIQUE (name)
);

-- 4) OBSERVATION_SESSION
CREATE TABLE IF NOT EXISTS observation_session (
    id                  INT AUTO_INCREMENT PRIMARY KEY,
    telescope_id        INT NULL, 
    instrument_id       INT NULL, 
    target_body_id      INT NULL, 
    mission_id          INT NULL, 
    start_time          DATETIME NOT NULL,
    end_time            DATETIME NULL,
    seeing_conditions   ENUM('Excellent','Good','Fair','Poor') DEFAULT 'Good',
    notes               TEXT
);

-- 5) DATA_SET
CREATE TABLE IF NOT EXISTS data_set (
    id                INT AUTO_INCREMENT PRIMARY KEY,
    name              VARCHAR(100) NOT NULL,
    mission_id        INT NULL,  
    observation_id    INT NULL,  
    data_description  TEXT,
    data_blob         LONGBLOB NULL,  
    collected_on      DATE NOT NULL,
    CONSTRAINT uq_data_set_name UNIQUE (name)
);

-- 6) RESEARCH_PAPER
CREATE TABLE IF NOT EXISTS research_paper (
    id             INT AUTO_INCREMENT PRIMARY KEY,
    title          VARCHAR(200) NOT NULL,
    abstract       TEXT,
    published_date DATE,
    doi            VARCHAR(100),  -- unique ID for academic papers
    project_id     INT NULL,     
    observatory_id INT NULL,      
    CONSTRAINT uq_paper_doi UNIQUE (doi),
    FULLTEXT INDEX idx_paper_ft (title, abstract)
);

-- 7) PAPER_CITATION
CREATE TABLE IF NOT EXISTS paper_citation (
    citing_paper_id    INT NOT NULL,  
    cited_paper_id     INT NOT NULL, 
    citation_date      DATE NOT NULL,
    PRIMARY KEY (citing_paper_id, cited_paper_id)
);

-- 8) GRANT
CREATE TABLE IF NOT EXISTS `grant` (
    id              INT AUTO_INCREMENT PRIMARY KEY,
    grant_number    VARCHAR(50) NOT NULL,
    agency_id       INT NOT NULL,   -- cross-schema reference: agency
    funding_amount  DECIMAL(15,2) NOT NULL DEFAULT 0,
    start_date      DATE NOT NULL,
    end_date        DATE NULL,
    status          ENUM('Proposed','Awarded','Active','Closed','Canceled') NOT NULL DEFAULT 'Proposed',
    CONSTRAINT uq_grant_number UNIQUE (grant_number)
);

-- 9) GRANT_RESEARCH_PROJECT
CREATE TABLE IF NOT EXISTS grant_research_project (
    grant_id          INT NOT NULL,
    research_project_id INT NOT NULL,  
    allocated_amount  DECIMAL(15,2),
    PRIMARY KEY (grant_id, research_project_id)
);

-- 10) INSTRUMENT_USAGE
CREATE TABLE IF NOT EXISTS instrument_usage (
    id               INT AUTO_INCREMENT PRIMARY KEY,
    instrument_id    INT NOT NULL,
    telescope_id     INT NULL,  
    spacecraft_id    INT NULL,   
    start_date       DATE NOT NULL,
    end_date         DATE NULL,
    usage_notes      TEXT
);

SET FOREIGN_KEY_CHECKS=1;


ALTER TABLE observatory
    ADD CONSTRAINT fk_obs_agency
        FOREIGN KEY (agency_id)
        REFERENCES agency(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    ADD CONSTRAINT fk_obs_launch_site
        FOREIGN KEY (launch_site_id)
        REFERENCES launch_site(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL;

ALTER TABLE telescope
    ADD CONSTRAINT fk_telescope_obs
        FOREIGN KEY (observatory_id)
        REFERENCES observatory(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE;

ALTER TABLE instrument
    ADD CONSTRAINT fk_instrument_telescope
        FOREIGN KEY (telescope_id)
        REFERENCES telescope(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL,
    ADD CONSTRAINT fk_instrument_spacecraft
        FOREIGN KEY (spacecraft_id)
        REFERENCES spacecraft(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL;

ALTER TABLE observation_session
    ADD CONSTRAINT fk_obs_sess_telescope
        FOREIGN KEY (telescope_id)
        REFERENCES telescope(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL,
    ADD CONSTRAINT fk_obs_sess_instrument
        FOREIGN KEY (instrument_id)
        REFERENCES instrument(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL,
    ADD CONSTRAINT fk_obs_sess_body
        FOREIGN KEY (target_body_id)
        REFERENCES celestial_body(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL,
    ADD CONSTRAINT fk_obs_sess_mission
        FOREIGN KEY (mission_id)
        REFERENCES mission(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL;

ALTER TABLE data_set
    ADD CONSTRAINT fk_dataset_mission
        FOREIGN KEY (mission_id)
        REFERENCES mission(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL,
    ADD CONSTRAINT fk_dataset_obs
        FOREIGN KEY (observation_id)
        REFERENCES observation_session(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL;

ALTER TABLE research_paper
    ADD CONSTRAINT fk_paper_proj
        FOREIGN KEY (project_id)
        REFERENCES research_project(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL,
    ADD CONSTRAINT fk_paper_obs
        FOREIGN KEY (observatory_id)
        REFERENCES observatory(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL;

ALTER TABLE paper_citation
    ADD CONSTRAINT fk_citation_citing
        FOREIGN KEY (citing_paper_id)
        REFERENCES research_paper(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    ADD CONSTRAINT fk_citation_cited
        FOREIGN KEY (cited_paper_id)
        REFERENCES research_paper(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE;

ALTER TABLE `grant`
    ADD CONSTRAINT fk_grant_agency
        FOREIGN KEY (agency_id)
        REFERENCES agency(id)
        ON UPDATE CASCADE
        ON DELETE RESTRICT;

ALTER TABLE grant_research_project
    ADD CONSTRAINT fk_grp_grant
        FOREIGN KEY (grant_id)
        REFERENCES `grant`(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    ADD CONSTRAINT fk_grp_project
        FOREIGN KEY (research_project_id)
        REFERENCES research_project(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE;

ALTER TABLE instrument_usage
    ADD CONSTRAINT fk_instrument_usage_instr
        FOREIGN KEY (instrument_id)
        REFERENCES instrument(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE,
    ADD CONSTRAINT fk_instrument_usage_telescope
        FOREIGN KEY (telescope_id)
        REFERENCES telescope(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL,
    ADD CONSTRAINT fk_instrument_usage_spacecraft
        FOREIGN KEY (spacecraft_id)
        REFERENCES spacecraft(id)
        ON UPDATE CASCADE
        ON DELETE SET NULL;

--
-- Sample Constraints & Triggers for Additional Complexity
--

ALTER TABLE telescope
    ADD CONSTRAINT chk_telescope_mirror CHECK (
        (telescope_type IN ('Optical','Infrared','UV','X-Ray') AND mirror_diameter_m > 0)
        OR (telescope_type = 'Radio')
        OR (telescope_type = 'Other')
    );

-- DELIMITER $$
CREATE TRIGGER trg_before_observation_session_insert
BEFORE INSERT ON observation_session
FOR EACH ROW
BEGIN
    IF NEW.telescope_id IS NOT NULL THEN
        IF (SELECT status FROM telescope WHERE id = NEW.telescope_id) = 'Retired' THEN
            SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'Cannot create observation session on a retired telescope.';
        END IF;
    END IF;
END;
-- DELIMITER ;

-- DELIMITER $$
CREATE PROCEDURE sp_award_grant(
    IN p_grant_id INT,
    IN p_start_date DATE,
    IN p_end_date DATE,
    IN p_amount DECIMAL(15,2)
)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION 
    BEGIN
        ROLLBACK;
        RESIGNAL;
    END;

    START TRANSACTION;
        UPDATE `grant`
        SET status = 'Awarded',
            start_date = p_start_date,
            end_date   = p_end_date,
            funding_amount = p_amount
        WHERE id = p_grant_id;
    COMMIT;
END;
-- DELIMITER ;

-- Example function: count citations of a research paper
-- DELIMITER $$
CREATE FUNCTION fn_paper_citation_count(p_paper_id INT)
RETURNS INT
DETERMINISTIC
READS SQL DATA
BEGIN
    DECLARE ccount INT;
    SELECT COUNT(*)
    INTO ccount
    FROM paper_citation
    WHERE cited_paper_id = p_paper_id;
    RETURN ccount;
END;
-- DELIMITER ;

--
-- Example View: v_cross_schema_projects
-- Joins a cosmic_research.grant with research_project 
-- to show cross-schema data
--
CREATE OR REPLACE VIEW v_cross_schema_projects AS
SELECT 
    g.id AS grant_id,
    g.grant_number,
    g.funding_amount,
    rp.id AS project_id,
    rp.title AS project_title,
    rp.description AS project_description,
    rp.start_date,
    rp.end_date,
    s_ag.name AS sponsoring_agency
FROM `grant` g
JOIN grant_research_project grp ON g.id = grp.grant_id
JOIN research_project rp ON grp.research_project_id = rp.id
LEFT JOIN agency s_ag ON g.agency_id = s_ag.id;

-- 5. Views

-- View: Mission Summary (combining details from multiple tables)
CREATE OR REPLACE VIEW v_mission_summary AS
SELECT 
    m.id AS mission_id,
    m.name AS mission_name,
    m.mission_type,
    m.status,
    m.launch_date,
    m.return_date,
    DATEDIFF(m.return_date, m.launch_date) AS duration_days,
    cb.name AS destination_name,
    ls.name AS launch_site_name,
    sc.name AS spacecraft_name,
    CONCAT(ascomm.first_name, ' ', ascomm.last_name) AS commander_name,
    ag.name AS primary_agency,
    -- subquery to count crew members on this mission
    (SELECT COUNT(*) FROM mission_crew mc WHERE mc.mission_id = m.id) AS crew_count
FROM mission m
LEFT JOIN astronaut ascomm ON m.commander_id = ascomm.id
LEFT JOIN spacecraft sc ON m.spacecraft_id = sc.id
LEFT JOIN celestial_body cb ON m.destination_id = cb.id
LEFT JOIN launch_site ls ON m.launch_site_id = ls.id
LEFT JOIN agency ag ON m.primary_agency_id = ag.id;

-- View: Astronaut Stats (missions count and total days in space per astronaut)
CREATE OR REPLACE VIEW v_astronaut_stats AS
SELECT 
    a.id AS astronaut_id,
    CONCAT(a.first_name, ' ', a.last_name) AS astronaut_name,
    a.nationality,
    a.status,
    ag.name AS agency_name,
    -- Count distinct missions (in case an astronaut had multiple roles in same mission, though our model prevents duplicates)
    IFNULL(COUNT(DISTINCT mc.mission_id), 0) AS mission_count,
    IFNULL(SUM(DATEDIFF(m.return_date, m.launch_date)), 0) AS total_days_in_space,
    fm.name AS first_mission_name
FROM astronaut a
LEFT JOIN agency ag ON a.agency_id = ag.id
LEFT JOIN mission fm ON a.first_mission_id = fm.id
LEFT JOIN mission_crew mc ON a.id = mc.astronaut_id
LEFT JOIN mission m ON mc.mission_id = m.id AND m.return_date IS NOT NULL  -- only count completed missions for days
GROUP BY a.id, a.first_name, a.last_name, a.nationality, a.status, ag.name, fm.name;




-- DELIMITER $$
-- Stored Function: Mission duration in days
CREATE FUNCTION mission_duration_days(p_mission_id INT)
RETURNS INT
DETERMINISTIC
READS SQL DATA
BEGIN
    DECLARE days INT;
    SELECT 
      CASE 
        WHEN return_date IS NOT NULL 
        THEN DATEDIFF(return_date, launch_date) 
        ELSE NULL 
      END
    INTO days
    FROM mission
    WHERE id = p_mission_id;
    RETURN days;
END;

-- DELIMITER ;

-- DELIMITER $$

CREATE FUNCTION astronaut_mission_count(p_astro_id INT)
RETURNS INT
DETERMINISTIC
READS SQL DATA
BEGIN
    DECLARE m_count INT;
    SELECT COUNT(*) INTO m_count
    FROM mission_crew
    WHERE astronaut_id = p_astro_id;
    RETURN IFNULL(m_count, 0);
END;

CREATE TRIGGER trg_before_mission_crew_insert
BEFORE INSERT ON mission_crew
FOR EACH ROW
BEGIN
    DECLARE current_commander INT;
    
    -- Ensure no duplicate commander assignment
    IF NEW.role = 'Commander' THEN 
        SELECT commander_id INTO current_commander 
        FROM mission 
        WHERE id = NEW.mission_id;
        IF current_commander IS NOT NULL AND current_commander != NEW.astronaut_id THEN
            SIGNAL SQLSTATE '45000'
                SET MESSAGE_TEXT = 'Cannot assign a second commander to the mission';
        END IF;
    END IF;
    
    -- Prevent crew assignment to unmanned missions
    IF (SELECT mission_type FROM mission WHERE id = NEW.mission_id) = 'Unmanned' THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'Cannot assign crew to an unmanned mission';
    END IF;
END;
-- Trigger: After inserting a mission crew record, update mission commander and astronaut first mission if needed
CREATE TRIGGER trg_after_mission_crew_insert
AFTER INSERT ON mission_crew
FOR EACH ROW
BEGIN
    -- If a commander was added to crew, ensure mission.commander_id is set to that astronaut
    IF NEW.role = 'Commander' THEN
        UPDATE mission
        SET commander_id = NEW.astronaut_id
        WHERE id = NEW.mission_id;
    END IF;
    -- If the astronaut has no first_mission recorded, set this mission as their first mission
    IF (SELECT first_mission_id FROM astronaut WHERE id = NEW.astronaut_id) IS NULL THEN
        UPDATE astronaut
        SET first_mission_id = NEW.mission_id
        WHERE id = NEW.astronaut_id;
    END IF;
END;

-- Trigger: Before inserting a mission, ensure unmanned missions have no commander (commander_id should be NULL if unmanned)
CREATE TRIGGER trg_before_mission_insert
BEFORE INSERT ON mission
FOR EACH ROW
BEGIN
    IF NEW.mission_type = 'Unmanned' AND NEW.commander_id IS NOT NULL THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'Unmanned missions cannot have a commander';
    END IF;
END;

-- Trigger: After updating a mission, log status/commander changes and update spacecraft status
CREATE TRIGGER trg_after_mission_update
AFTER UPDATE ON mission
FOR EACH ROW
BEGIN
    -- Log status change
    IF NEW.status <> OLD.status THEN
        INSERT INTO mission_log(mission_id, event)
        VALUES (NEW.id, CONCAT('Mission status changed from ', OLD.status, ' to ', NEW.status));
        -- Update spacecraft status based on mission status changes
        IF NEW.status = 'Active' THEN
            UPDATE spacecraft
            SET status = 'In Mission'
            WHERE id = NEW.spacecraft_id;
        ELSEIF NEW.status = 'Completed' THEN
            UPDATE spacecraft
            SET status = 'Operational',
                last_mission_id = NEW.id
            WHERE id = NEW.spacecraft_id;
        ELSEIF NEW.status IN ('Failed','Aborted') THEN
            UPDATE spacecraft
            SET status = 'Lost',
                last_mission_id = NEW.id
            WHERE id = NEW.spacecraft_id;
        END IF;
    END IF;
    -- Log commander change
    IF NEW.commander_id <> OLD.commander_id THEN
        IF NEW.commander_id IS NULL THEN
            INSERT INTO mission_log(mission_id, event)
            VALUES (NEW.id, 'Mission commander unassigned');
        ELSE 
            -- Fetch new commander's name for the log
            INSERT INTO mission_log(mission_id, event)
            VALUES (NEW.id, CONCAT('Commander changed to astronaut ID ', NEW.commander_id));
        END IF;
    END IF;
END;

-- DELIMITER ;
-- 3. Stored Procedures and Functions

-- DELIMITER $$

-- Stored Procedure: Assign an astronaut to a mission with a given role (with transaction)
CREATE PROCEDURE assign_astronaut_to_mission(
    IN p_mission_id INT,
    IN p_astronaut_id INT,
    IN p_role VARCHAR(20)
)
BEGIN
    -- Error handler: rollback on any error
    DECLARE EXIT HANDLER FOR SQLEXCEPTION 
    BEGIN 
        ROLLBACK;
        RESIGNAL;  
    END;
    START TRANSACTION;
        -- Insert crew assignment
        INSERT INTO mission_crew(mission_id, astronaut_id, role)
        VALUES (p_mission_id, p_astronaut_id, p_role);
        -- If role is Commander, update mission's commander_id (assign this astronaut as mission commander)
        IF UPPER(p_role) = 'COMMANDER' THEN
            UPDATE mission 
            SET commander_id = p_astronaut_id
            WHERE id = p_mission_id;
        END IF;
        -- If astronaut has no first mission recorded, set it to this mission
        IF (SELECT first_mission_id FROM astronaut WHERE id = p_astronaut_id) IS NULL THEN
            UPDATE astronaut
            SET first_mission_id = p_mission_id
            WHERE id = p_astronaut_id;
        END IF;
    COMMIT;
END;
