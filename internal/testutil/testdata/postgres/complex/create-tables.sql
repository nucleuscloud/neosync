CREATE SCHEMA IF NOT EXISTS scientific_data;
CREATE SCHEMA IF NOT EXISTS space_mission;
set search_path to space_mission;

CREATE TABLE IF NOT EXISTS astronauts (
    astronaut_id       SERIAL PRIMARY KEY,
    name               VARCHAR(100) NOT NULL,
    email              VARCHAR(100) UNIQUE NOT NULL,
    manager_astronaut_id INTEGER,
    mentor_astronaut_id  INTEGER
);

CREATE TABLE IF NOT EXISTS missions (
    mission_id          SERIAL PRIMARY KEY,
    name                VARCHAR(100) NOT NULL,
    description         TEXT,
    status              VARCHAR(50) NOT NULL DEFAULT 'Pending',
    lead_astronaut_id   INTEGER,
    client_astronaut_id INTEGER
);

CREATE TABLE IF NOT EXISTS objectives (
    objective_id            SERIAL PRIMARY KEY,
    title                   VARCHAR(200) NOT NULL,
    description             TEXT,
    status                  VARCHAR(50),
    mission_id              INTEGER,
    assignee_astronaut_id   INTEGER,
    reviewer_astronaut_id   INTEGER
);

CREATE TABLE IF NOT EXISTS capabilities (
    capability_id SERIAL PRIMARY KEY,
    name          VARCHAR(100) NOT NULL,
    category      VARCHAR(100)
);


CREATE TABLE IF NOT EXISTS astronaut_capabilities (
    astronaut_capability_id SERIAL PRIMARY KEY,
    astronaut_id            INTEGER,
    capability_id           INTEGER,
    proficiency_level       INTEGER CHECK (proficiency_level BETWEEN 1 AND 5)
);

CREATE TABLE IF NOT EXISTS transmissions (
    transmission_id     SERIAL PRIMARY KEY,
    content             TEXT NOT NULL,
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    astronaut_id        INTEGER,
    objective_id        INTEGER,
    mission_id          INTEGER,
    parent_transmission_id INTEGER
);

CREATE TABLE IF NOT EXISTS payloads (
    payload_id               SERIAL PRIMARY KEY,
    file_name                VARCHAR(255) NOT NULL,
    file_path                VARCHAR(255) NOT NULL,
    uploaded_by_astronaut_id INTEGER,
    objective_id             INTEGER,
    mission_id               INTEGER,
    transmission_id          INTEGER
);

CREATE TABLE IF NOT EXISTS crew_assignments (
    crew_assignment_id SERIAL PRIMARY KEY,
    astronaut_id       INTEGER NOT NULL,
    mission_id         INTEGER NOT NULL,
    role               VARCHAR(100)
);


CREATE TABLE IF NOT EXISTS mission_logs (
    log_id      SERIAL PRIMARY KEY,
    object_type VARCHAR(100),
    object_id   INTEGER,
    action      VARCHAR(200),
    "timestamp" TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS crews (
    crew_id          SERIAL PRIMARY KEY,
    crew_name        VARCHAR(100) NOT NULL,
    lead_astronaut_id INTEGER,            
    parent_crew_id    INTEGER             
);

CREATE TABLE IF NOT EXISTS crew_missions (
    crew_mission_id SERIAL PRIMARY KEY,
    crew_id         INTEGER NOT NULL,
    mission_id      INTEGER NOT NULL,
    notes           VARCHAR(255)
);


CREATE TABLE IF NOT EXISTS supplies (
    supply_id            SERIAL PRIMARY KEY,
    mission_id           INTEGER,         
    bill_to_astronaut_id INTEGER,         
    owner_astronaut_id   INTEGER,         
    total_amount         DECIMAL(12, 2),
    status               VARCHAR(50) NOT NULL DEFAULT 'Open',
    created_at           TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS supply_items (
    supply_item_id SERIAL PRIMARY KEY,
    supply_id      INTEGER NOT NULL, 
    objective_id   INTEGER,          
    description    VARCHAR(255) NOT NULL,
    quantity       INTEGER NOT NULL DEFAULT 1,
    unit_price     DECIMAL(10, 2) NOT NULL DEFAULT 0
);


ALTER TABLE astronauts
    ADD CONSTRAINT fk_astronaut_manager
        FOREIGN KEY (manager_astronaut_id) REFERENCES astronauts(astronaut_id),
    ADD CONSTRAINT fk_astronaut_mentor
        FOREIGN KEY (mentor_astronaut_id) REFERENCES astronauts(astronaut_id);

ALTER TABLE missions
    ADD CONSTRAINT fk_mission_lead
        FOREIGN KEY (lead_astronaut_id) REFERENCES astronauts(astronaut_id),
    ADD CONSTRAINT fk_mission_client
        FOREIGN KEY (client_astronaut_id) REFERENCES astronauts(astronaut_id);

ALTER TABLE objectives
    ADD CONSTRAINT fk_objective_mission
        FOREIGN KEY (mission_id) REFERENCES missions(mission_id),
    ADD CONSTRAINT fk_objective_assignee
        FOREIGN KEY (assignee_astronaut_id) REFERENCES astronauts(astronaut_id),
    ADD CONSTRAINT fk_objective_reviewer
        FOREIGN KEY (reviewer_astronaut_id) REFERENCES astronauts(astronaut_id);

ALTER TABLE astronaut_capabilities
    ADD CONSTRAINT fk_astronaut_capability_astronaut
        FOREIGN KEY (astronaut_id) REFERENCES astronauts(astronaut_id),
    ADD CONSTRAINT fk_astronaut_capability_capability
        FOREIGN KEY (capability_id) REFERENCES capabilities(capability_id);

ALTER TABLE transmissions
    ADD CONSTRAINT fk_transmission_astronaut
        FOREIGN KEY (astronaut_id) REFERENCES astronauts(astronaut_id),
    ADD CONSTRAINT fk_transmission_objective
        FOREIGN KEY (objective_id) REFERENCES objectives(objective_id),
    ADD CONSTRAINT fk_transmission_mission
        FOREIGN KEY (mission_id) REFERENCES missions(mission_id),
    ADD CONSTRAINT fk_transmission_parent
        FOREIGN KEY (parent_transmission_id) REFERENCES transmissions(transmission_id);

ALTER TABLE payloads
    ADD CONSTRAINT fk_payload_astronaut
        FOREIGN KEY (uploaded_by_astronaut_id) REFERENCES astronauts(astronaut_id),
    ADD CONSTRAINT fk_payload_objective
        FOREIGN KEY (objective_id) REFERENCES objectives(objective_id),
    ADD CONSTRAINT fk_payload_mission
        FOREIGN KEY (mission_id) REFERENCES missions(mission_id),
    ADD CONSTRAINT fk_payload_transmission
        FOREIGN KEY (transmission_id) REFERENCES transmissions(transmission_id);

ALTER TABLE crew_assignments
    ADD CONSTRAINT fk_crewassign_astronaut
        FOREIGN KEY (astronaut_id) REFERENCES astronauts(astronaut_id),
    ADD CONSTRAINT fk_crewassign_mission
        FOREIGN KEY (mission_id) REFERENCES missions(mission_id);

ALTER TABLE crews
    ADD CONSTRAINT fk_crew_lead
        FOREIGN KEY (lead_astronaut_id) REFERENCES astronauts(astronaut_id),
    ADD CONSTRAINT fk_crew_parent
        FOREIGN KEY (parent_crew_id) REFERENCES crews(crew_id);

ALTER TABLE crew_missions
    ADD CONSTRAINT fk_crewmissions_crew
        FOREIGN KEY (crew_id) REFERENCES crews(crew_id),
    ADD CONSTRAINT fk_crewmissions_mission
        FOREIGN KEY (mission_id) REFERENCES missions(mission_id);

ALTER TABLE supplies
    ADD CONSTRAINT fk_supply_mission
        FOREIGN KEY (mission_id) REFERENCES missions(mission_id),
    ADD CONSTRAINT fk_supply_bill_to
        FOREIGN KEY (bill_to_astronaut_id) REFERENCES astronauts(astronaut_id),
    ADD CONSTRAINT fk_supply_owner
        FOREIGN KEY (owner_astronaut_id) REFERENCES astronauts(astronaut_id);

ALTER TABLE supply_items
    ADD CONSTRAINT fk_supply_item_supply
        FOREIGN KEY (supply_id) REFERENCES supplies(supply_id),
    ADD CONSTRAINT fk_supply_item_objective
        FOREIGN KEY (objective_id) REFERENCES objectives(objective_id);



CREATE TABLE IF NOT EXISTS spacecraft_class (
    class_id SERIAL PRIMARY KEY,
    class_name VARCHAR(100) NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS spacecraft (
    spacecraft_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    class_id INTEGER NOT NULL,
    launch_date DATE,
    status VARCHAR(50),
    CONSTRAINT fk_spacecraft_class FOREIGN KEY (class_id) 
        REFERENCES spacecraft_class(class_id)
);

CREATE TABLE IF NOT EXISTS spacecraft_module (
    module_id SERIAL PRIMARY KEY,
    spacecraft_id INTEGER NOT NULL,
    module_name VARCHAR(100) NOT NULL,
    purpose TEXT,
    installation_date DATE,
    CONSTRAINT fk_module_spacecraft FOREIGN KEY (spacecraft_id) 
        REFERENCES spacecraft(spacecraft_id)
);

CREATE TABLE IF NOT EXISTS module_component (
    component_id SERIAL PRIMARY KEY,
    module_id INTEGER NOT NULL,
    component_name VARCHAR(100) NOT NULL,
    manufacturer VARCHAR(100),
    serial_number VARCHAR(50),
    installation_date DATE,
    CONSTRAINT fk_component_module FOREIGN KEY (module_id) 
        REFERENCES spacecraft_module(module_id)
);


CREATE TABLE IF NOT EXISTS equipment (
    equipment_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50),
    weight_kg DECIMAL(10,2)
);

CREATE TABLE IF NOT EXISTS mission_equipment (
    mission_id INTEGER NOT NULL,
    equipment_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    assignment_date DATE NOT NULL,
    PRIMARY KEY (mission_id, equipment_id),
    CONSTRAINT fk_mission_equipment_mission FOREIGN KEY (mission_id) 
        REFERENCES missions(mission_id),
    CONSTRAINT fk_mission_equipment_equipment FOREIGN KEY (equipment_id) 
        REFERENCES equipment(equipment_id)
);

CREATE TABLE IF NOT EXISTS equipment_maintenance (
    maintenance_id SERIAL PRIMARY KEY,
    mission_id INTEGER NOT NULL,
    equipment_id INTEGER NOT NULL,
    maintenance_date DATE NOT NULL,
    performed_by_astronaut_id INTEGER,
    description TEXT,
    CONSTRAINT fk_maintenance_mission_equipment FOREIGN KEY (mission_id, equipment_id) 
        REFERENCES mission_equipment(mission_id, equipment_id),
    CONSTRAINT fk_maintenance_astronaut FOREIGN KEY (performed_by_astronaut_id) 
        REFERENCES astronauts(astronaut_id)
);

CREATE TABLE IF NOT EXISTS training_courses (
    course_id SERIAL PRIMARY KEY,
    course_name VARCHAR(100) NOT NULL,
    duration_hours INTEGER,
    description TEXT
);

CREATE TABLE IF NOT EXISTS course_prerequisites (
    prerequisite_id SERIAL PRIMARY KEY,
    course_id INTEGER NOT NULL,
    required_course_id INTEGER NOT NULL,
    CONSTRAINT fk_prerequisite_course FOREIGN KEY (course_id) 
        REFERENCES training_courses(course_id),
    CONSTRAINT fk_required_course FOREIGN KEY (required_course_id) 
        REFERENCES training_courses(course_id),
    CONSTRAINT check_not_self_reference CHECK (course_id != required_course_id)
);

CREATE TABLE IF NOT EXISTS certifications (
    certification_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    issuing_authority VARCHAR(100),
    valid_years INTEGER
);

CREATE TABLE IF NOT EXISTS astronaut_certifications (
    astronaut_id INTEGER NOT NULL,
    certification_id INTEGER NOT NULL,
    issue_date DATE NOT NULL,
    expiry_date DATE,
    certificate_number VARCHAR(50),
    PRIMARY KEY (astronaut_id, certification_id),
    CONSTRAINT fk_astro_cert_astronaut FOREIGN KEY (astronaut_id) 
        REFERENCES astronauts(astronaut_id),
    CONSTRAINT fk_astro_cert_certification FOREIGN KEY (certification_id) 
        REFERENCES certifications(certification_id)
);

CREATE TABLE IF NOT EXISTS certification_requirements (
    requirement_id SERIAL PRIMARY KEY,
    certification_id INTEGER NOT NULL,
    required_certification_id INTEGER NOT NULL,
    CONSTRAINT fk_cert_requirement FOREIGN KEY (certification_id) 
        REFERENCES certifications(certification_id),
    CONSTRAINT fk_required_cert FOREIGN KEY (required_certification_id) 
        REFERENCES certifications(certification_id),
    CONSTRAINT check_not_self_reference CHECK (certification_id != required_certification_id)
);




CREATE TABLE IF NOT EXISTS mission_logs_extended (
    log_id SERIAL PRIMARY KEY,
    mission_id INTEGER NOT NULL,
    astronaut_id INTEGER, 
    log_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    log_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20),
    message TEXT,
    CONSTRAINT fk_mission_log_mission FOREIGN KEY (mission_id) 
        REFERENCES missions(mission_id),
    CONSTRAINT fk_mission_log_astronaut FOREIGN KEY (astronaut_id) 
        REFERENCES astronauts(astronaut_id),
    CONSTRAINT uq_mission_log UNIQUE (mission_id, astronaut_id, log_date)
);

CREATE TABLE IF NOT EXISTS communication_channels (
    channel_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    frequency VARCHAR(50),
    encryption_level VARCHAR(50)
);

CREATE TABLE IF NOT EXISTS mission_communications (
    mission_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    priority INTEGER NOT NULL,
    is_backup BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (mission_id, channel_id),
    CONSTRAINT fk_mission_comm_mission FOREIGN KEY (mission_id) 
        REFERENCES missions(mission_id),
    CONSTRAINT fk_mission_comm_channel FOREIGN KEY (channel_id) 
        REFERENCES communication_channels(channel_id)
);

CREATE TABLE IF NOT EXISTS message_logs (
    log_id SERIAL PRIMARY KEY,
    mission_id INTEGER NOT NULL,
    channel_id INTEGER, 
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sender_astronaut_id INTEGER,
    content TEXT,
    is_encrypted BOOLEAN DEFAULT FALSE,
    CONSTRAINT fk_message_mission_channel FOREIGN KEY (mission_id, channel_id) 
        REFERENCES mission_communications(mission_id, channel_id) ON DELETE SET NULL,
    CONSTRAINT fk_message_astronaut FOREIGN KEY (sender_astronaut_id) 
        REFERENCES astronauts(astronaut_id)
);

CREATE TABLE IF NOT EXISTS events (
    event_id SERIAL PRIMARY KEY,
    event_time TIMESTAMP NOT NULL,
    severity VARCHAR(20) CHECK (severity IN ('Info', 'Warning', 'Critical', 'Emergency')),
    title VARCHAR(200) NOT NULL,
    description TEXT
);

-- not currently supported
-- CREATE TABLE IF NOT EXISTS system_events (
--     system_name VARCHAR(100) NOT NULL,
--     component_id INTEGER,
--     error_code VARCHAR(50),
--     resolved BOOLEAN DEFAULT FALSE
-- ) INHERITS (events);

-- CREATE TABLE IF NOT EXISTS astronaut_events (
--     astronaut_id INTEGER NOT NULL REFERENCES astronauts(astronaut_id),
--     location VARCHAR(100),
--     vital_signs JSONB,
--     mission_id INTEGER REFERENCES missions(mission_id)
-- ) INHERITS (events);

-- CREATE TABLE IF NOT EXISTS mission_events (
--     mission_id INTEGER NOT NULL REFERENCES missions(mission_id),
--     milestone VARCHAR(100),
--     success_rating INTEGER CHECK (success_rating BETWEEN 1 AND 10),
--     affected_objectives INTEGER[]
-- ) INHERITS (events);

-- CREATE TABLE telemetry (
--     telemetry_id BIGSERIAL,
--     timestamp TIMESTAMP NOT NULL,
--     spacecraft_id INTEGER NOT NULL REFERENCES spacecraft(spacecraft_id),
--     sensor_type VARCHAR(50) NOT NULL,
--     reading NUMERIC(12,4) NOT NULL,
--     unit VARCHAR(20),
--     coordinates POINT,
--     is_anomaly BOOLEAN DEFAULT FALSE,
--     PRIMARY KEY (telemetry_id, timestamp)
-- ) PARTITION BY RANGE (timestamp);

-- CREATE TABLE telemetry_2023 PARTITION OF telemetry
--     FOR VALUES FROM ('2023-01-01') TO ('2024-01-01');
-- CREATE TABLE telemetry_2024 PARTITION OF telemetry
--     FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
-- CREATE TABLE telemetry_2025 PARTITION OF telemetry
--     FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');


CREATE TABLE IF NOT EXISTS comments (
    comment_id SERIAL PRIMARY KEY,
    reference_type VARCHAR(50) NOT NULL,
    reference_id INTEGER NOT NULL,
    author_astronaut_id INTEGER REFERENCES astronauts(astronaut_id),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    CONSTRAINT uq_polymorphic_ref UNIQUE (reference_type, reference_id, comment_id)
);

CREATE TABLE IF NOT EXISTS tags (
    tag_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(50),
    created_by_astronaut_id INTEGER REFERENCES astronauts(astronaut_id)
);

CREATE TABLE IF NOT EXISTS taggables (
    tag_id INTEGER NOT NULL REFERENCES tags(tag_id),
    taggable_type VARCHAR(50) NOT NULL,
    taggable_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by_astronaut_id INTEGER REFERENCES astronauts(astronaut_id),
    PRIMARY KEY (tag_id, taggable_type, taggable_id)
);



CREATE TABLE IF NOT EXISTS scientific_data.experiments (
    experiment_id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    lead_scientist_astronaut_id INTEGER REFERENCES space_mission.astronauts(astronaut_id),
    start_date DATE,
    end_date DATE,
    status VARCHAR(50),
    mission_id INTEGER REFERENCES space_mission.missions(mission_id)
);

CREATE TABLE IF NOT EXISTS scientific_data.samples (
    sample_id SERIAL PRIMARY KEY,
    experiment_id INTEGER REFERENCES scientific_data.experiments(experiment_id),
    collection_date TIMESTAMP,
    location VARCHAR(100),
    collected_by_astronaut_id INTEGER REFERENCES space_mission.astronauts(astronaut_id),
    sample_type VARCHAR(50),
    storage_conditions VARCHAR(100),
    mission_id INTEGER REFERENCES space_mission.missions(mission_id),
    spacecraft_id INTEGER REFERENCES space_mission.spacecraft(spacecraft_id)
);

CREATE TABLE IF NOT EXISTS scientific_data.measurements (
    measurement_id SERIAL PRIMARY KEY,
    sample_id INTEGER REFERENCES scientific_data.samples(sample_id),
    parameter VARCHAR(100) NOT NULL,
    value NUMERIC(15,6),
    unit VARCHAR(30),
    measured_at TIMESTAMP,
    measured_by_astronaut_id INTEGER REFERENCES space_mission.astronauts(astronaut_id),
    instrument VARCHAR(100),
    confidence_level DECIMAL(5,2),
    is_verified BOOLEAN DEFAULT FALSE
);


CREATE TABLE IF NOT EXISTS space_mission.mission_experiments (
    mission_id INTEGER REFERENCES space_mission.missions(mission_id),
    experiment_id INTEGER REFERENCES scientific_data.experiments(experiment_id),
    priority INTEGER,
    time_allocation_hours INTEGER,
    resources_allocated JSONB,
    notes TEXT,
    PRIMARY KEY (mission_id, experiment_id)
);


CREATE OR REPLACE VIEW space_mission.astronaut_experiment_assignments AS
SELECT 
    a.astronaut_id,
    a.name AS astronaut_name,
    e.experiment_id,
    e.name AS experiment_name,
    m.mission_id,
    m.name AS mission_name,
    s.sample_id,
    s.sample_type,
    s.collection_date
FROM 
    space_mission.astronauts a
JOIN 
    scientific_data.experiments e ON a.astronaut_id = e.lead_scientist_astronaut_id
JOIN 
    space_mission.missions m ON e.mission_id = m.mission_id
LEFT JOIN 
    scientific_data.samples s ON e.experiment_id = s.experiment_id AND s.collected_by_astronaut_id = a.astronaut_id;


CREATE MATERIALIZED VIEW space_mission.mission_performance_metrics AS
SELECT 
    m.mission_id,
    m.name AS mission_name,
    COUNT(DISTINCT o.objective_id) AS total_objectives,
    COUNT(DISTINCT o.objective_id) FILTER (WHERE o.status = 'Completed') AS completed_objectives,
    COUNT(DISTINCT ca.astronaut_id) AS crew_size,
    COUNT(DISTINCT ac.capability_id) AS team_capabilities,
    AVG(ac.proficiency_level) AS avg_proficiency,
    -- COUNT(DISTINCT me.event_id) AS mission_events,
    -- COUNT(DISTINCT me.event_id) FILTER (WHERE me.severity = 'Critical') AS critical_events,
    COALESCE(SUM(si.quantity * si.unit_price), 0) AS total_supply_cost
FROM 
    space_mission.missions m
LEFT JOIN 
    space_mission.objectives o ON m.mission_id = o.mission_id
LEFT JOIN 
    space_mission.crew_assignments ca ON m.mission_id = ca.mission_id
LEFT JOIN 
    space_mission.astronaut_capabilities ac ON ca.astronaut_id = ac.astronaut_id
-- LEFT JOIN 
--     space_mission.mission_events me ON m.mission_id = me.mission_id
LEFT JOIN 
    space_mission.supplies s ON m.mission_id = s.mission_id
LEFT JOIN 
    space_mission.supply_items si ON s.supply_id = si.supply_id
GROUP BY 
    m.mission_id, m.name;


-- CREATE OR REPLACE VIEW space_mission.high_performing_missions AS
-- SELECT 
--     mission_id,
--     mission_name,
--     completed_objectives::FLOAT / NULLIF(total_objectives, 0) AS completion_rate,
--     crew_size,
--     team_capabilities,
--     avg_proficiency,
--     critical_events,
--     total_supply_cost
-- FROM 
--     space_mission.mission_performance_metrics
-- WHERE 
--     (completed_objectives::FLOAT / NULLIF(total_objectives, 0)) > 0.7
-- AND 
--     critical_events < 3;

CREATE TABLE IF NOT EXISTS space_mission.mission_parameters (
    parameter_id SERIAL PRIMARY KEY,
    mission_id INTEGER REFERENCES space_mission.missions(mission_id),
    name VARCHAR(100) NOT NULL,
    numeric_values NUMERIC[] DEFAULT '{}',
    string_values TEXT[] DEFAULT '{}',
    json_config JSONB DEFAULT '{}'::jsonb,
    applicable_spacecraft INTEGER[] DEFAULT '{}',
    tags VARCHAR[] DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version INTEGER DEFAULT 1
);

CREATE INDEX idx_mission_params_spacecraft ON space_mission.mission_parameters USING GIN (applicable_spacecraft);
CREATE INDEX idx_mission_params_tags ON space_mission.mission_parameters USING GIN (tags);
CREATE INDEX idx_mission_params_json ON space_mission.mission_parameters USING GIN (json_config);

CREATE TABLE IF NOT EXISTS space_mission.skill_groups (
    group_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_group_id INTEGER REFERENCES space_mission.skill_groups(group_id),
    required_for_role VARCHAR[] DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS space_mission.capability_skill_groups (
    capability_id INTEGER REFERENCES space_mission.capabilities(capability_id),
    group_id INTEGER REFERENCES space_mission.skill_groups(group_id),
    importance_level INTEGER CHECK (importance_level BETWEEN 1 AND 5),
    PRIMARY KEY (capability_id, group_id)
);

CREATE TABLE IF NOT EXISTS space_mission.mission_required_skill_groups (
    mission_id INTEGER REFERENCES space_mission.missions(mission_id),
    group_id INTEGER REFERENCES space_mission.skill_groups(group_id),
    minimum_proficiency INTEGER CHECK (minimum_proficiency BETWEEN 1 AND 5),
    role VARCHAR(100),
    is_critical BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (mission_id, group_id, role)
);


CREATE TABLE IF NOT EXISTS space_mission.equipment_compatibility (
    primary_equipment_id INTEGER REFERENCES space_mission.equipment(equipment_id),
    compatible_equipment_id INTEGER REFERENCES space_mission.equipment(equipment_id),
    compatibility_level INTEGER CHECK (compatibility_level BETWEEN 0 AND 5),
    notes TEXT,
    PRIMARY KEY (primary_equipment_id, compatible_equipment_id),
    CHECK (primary_equipment_id != compatible_equipment_id)
);

CREATE DOMAIN space_mission.positive_decimal AS DECIMAL(12,2) 
    CHECK (VALUE > 0);

CREATE DOMAIN space_mission.email_address AS VARCHAR(255)
    CHECK (VALUE ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$');

CREATE DOMAIN space_mission.status_type AS VARCHAR(50)
    CHECK (VALUE IN ('Planned', 'In Progress', 'Completed', 'Cancelled', 'On Hold'));


CREATE TYPE space_mission.coordinate AS (
    longitude DECIMAL(9,6),
    latitude DECIMAL(9,6),
    altitude DECIMAL(9,2)
);

CREATE TYPE space_mission.vital_signs AS (
    heart_rate INTEGER,
    blood_pressure VARCHAR(10),
    temperature DECIMAL(4,1),
    oxygen_saturation DECIMAL(4,1),
    timestamp TIMESTAMP
);

ALTER TABLE space_mission.supplies
    ALTER COLUMN total_amount TYPE space_mission.positive_decimal;

ALTER TABLE space_mission.astronauts
    ALTER COLUMN email TYPE space_mission.email_address;

CREATE TABLE IF NOT EXISTS space_mission.astronaut_vitals (
    vital_id SERIAL PRIMARY KEY,
    astronaut_id INTEGER REFERENCES space_mission.astronauts(astronaut_id),
    mission_id INTEGER REFERENCES space_mission.missions(mission_id),
    location space_mission.coordinate,
    vital_data space_mission.vital_signs,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_anomalous BOOLEAN DEFAULT FALSE
);

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS space_mission.mission_status_history (
    history_id SERIAL PRIMARY KEY,
    mission_id INTEGER REFERENCES space_mission.missions(mission_id),
    status VARCHAR(50) NOT NULL,
    effective_from TIMESTAMP NOT NULL,
    effective_to TIMESTAMP,
    modified_by_astronaut_id INTEGER REFERENCES space_mission.astronauts(astronaut_id),
    reason TEXT,
    CONSTRAINT mission_status_history_excl EXCLUDE USING GIST (
        mission_id WITH =,
        daterange(effective_from::date, COALESCE(effective_to::date, 'infinity'::date), '[)') WITH &&
    )
);

CREATE TABLE IF NOT EXISTS space_mission.equipment_status_history (
    history_id SERIAL PRIMARY KEY,
    equipment_id INTEGER REFERENCES space_mission.equipment(equipment_id),
    mission_id INTEGER REFERENCES space_mission.missions(mission_id),
    status VARCHAR(50) NOT NULL,
    effective_from TIMESTAMP NOT NULL,
    effective_to TIMESTAMP,
    location space_mission.coordinate,
    condition_rating INTEGER CHECK (condition_rating BETWEEN 1 AND 10),
    notes TEXT,
    CONSTRAINT equipment_status_history_excl EXCLUDE USING GIST (
        equipment_id WITH =,
        mission_id WITH =,
        daterange(effective_from::date, COALESCE(effective_to::date, 'infinity'::date), '[)') WITH &&
    )
);

CREATE TABLE IF NOT EXISTS space_mission.astronaut_role_history (
    history_id SERIAL PRIMARY KEY,
    astronaut_id INTEGER REFERENCES space_mission.astronauts(astronaut_id),
    role VARCHAR(100) NOT NULL,
    mission_id INTEGER REFERENCES space_mission.missions(mission_id),
    effective_from TIMESTAMP NOT NULL,
    effective_to TIMESTAMP,
    supervisor_astronaut_id INTEGER REFERENCES space_mission.astronauts(astronaut_id),
    performance_rating INTEGER CHECK (performance_rating BETWEEN 1 AND 5),
    notes TEXT,
    CONSTRAINT astronaut_role_history_excl EXCLUDE USING GIST (
        astronaut_id WITH =,
        mission_id WITH =,
        daterange(effective_from::date, COALESCE(effective_to::date, 'infinity'::date), '[)') WITH &&
    )
);

CREATE OR REPLACE FUNCTION space_mission.find_certification_path(
    p_astronaut_id INTEGER,
    p_target_certification_id INTEGER
) RETURNS TABLE (
    path_level INTEGER,
    certification_id INTEGER,
    certification_name VARCHAR,
    has_certification BOOLEAN,
    issue_date DATE,
    expiry_date DATE
) AS $$
BEGIN
    RETURN QUERY WITH RECURSIVE cert_path AS (
        
        SELECT 
            0 AS level, 
            c.certification_id,
            c.name,
            EXISTS (
                SELECT 1 FROM space_mission.astronaut_certifications ac 
                WHERE ac.astronaut_id = p_astronaut_id AND ac.certification_id = c.certification_id
            ) AS has_cert,
            (
                SELECT ac.issue_date FROM space_mission.astronaut_certifications ac 
                WHERE ac.astronaut_id = p_astronaut_id AND ac.certification_id = c.certification_id
            ) AS issue_date,
            (
                SELECT ac.expiry_date FROM space_mission.astronaut_certifications ac 
                WHERE ac.astronaut_id = p_astronaut_id AND ac.certification_id = c.certification_id
            ) AS expiry_date
        FROM 
            space_mission.certifications c
        WHERE 
            c.certification_id = p_target_certification_id
        
        UNION ALL
        
        
        SELECT 
            cp.level + 1, 
            c.certification_id,
            c.name,
            EXISTS (
                SELECT 1 FROM space_mission.astronaut_certifications ac 
                WHERE ac.astronaut_id = p_astronaut_id AND ac.certification_id = c.certification_id
            ) AS has_cert,
            (
                SELECT ac.issue_date FROM space_mission.astronaut_certifications ac 
                WHERE ac.astronaut_id = p_astronaut_id AND ac.certification_id = c.certification_id
            ) AS issue_date,
            (
                SELECT ac.expiry_date FROM space_mission.astronaut_certifications ac 
                WHERE ac.astronaut_id = p_astronaut_id AND ac.certification_id = c.certification_id
            ) AS expiry_date
        FROM 
            cert_path cp
        JOIN 
            space_mission.certification_requirements cr ON cp.certification_id = cr.certification_id
        JOIN 
            space_mission.certifications c ON cr.required_certification_id = c.certification_id
        WHERE 
            cp.level < 10  
    )
    SELECT 
        level AS path_level,
        certification_id,
        name AS certification_name,
        has_cert AS has_certification,
        issue_date,
        expiry_date
    FROM 
        cert_path
    ORDER BY 
        level, certification_id;
END;
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION space_mission.find_equipment_dependencies(
    p_equipment_id INTEGER,
    p_compatibility_threshold INTEGER DEFAULT 3
) RETURNS TABLE (
    depth INTEGER,
    equipment_id INTEGER,
    equipment_name VARCHAR,
    compatibility_path TEXT,
    avg_compatibility DECIMAL
) AS $$
BEGIN
    RETURN QUERY WITH RECURSIVE equip_deps AS (
        
        SELECT 
            0 AS depth, 
            e.equipment_id,
            e.name AS equipment_name,
            e.name::TEXT AS path,
            5.0 AS avg_compatibility
        FROM 
            space_mission.equipment e
        WHERE 
            e.equipment_id = p_equipment_id
        
        UNION ALL
        
        
        SELECT 
            ed.depth + 1, 
            e.equipment_id,
            e.name,
            ed.path || ' -> ' || e.name,
            (ed.avg_compatibility * ed.depth + ec.compatibility_level)::DECIMAL / (ed.depth + 1)
        FROM 
            equip_deps ed
        JOIN 
            space_mission.equipment_compatibility ec ON ed.equipment_id = ec.primary_equipment_id
        JOIN 
            space_mission.equipment e ON ec.compatible_equipment_id = e.equipment_id
        WHERE 
            ed.depth < 5  
        AND 
            ec.compatibility_level >= p_compatibility_threshold
        AND 
            NOT (e.equipment_id = ANY(string_to_array(ed.path, ' -> ')::INTEGER[]))  
    )
    SELECT 
        depth,
        equipment_id,
        equipment_name,
        path AS compatibility_path,
        avg_compatibility
    FROM 
        equip_deps
    ORDER BY 
        depth, equipment_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION space_mission.update_certification_status()
RETURNS TRIGGER AS $$
BEGIN
    
    IF NEW.expiry_date <= CURRENT_DATE AND OLD.expiry_date > CURRENT_DATE THEN
        UPDATE space_mission.astronaut_certifications ac
        SET expiry_date = CURRENT_DATE
        FROM space_mission.certification_requirements cr
        WHERE cr.required_certification_id = NEW.certification_id
        AND cr.certification_id = ac.certification_id
        AND ac.astronaut_id = NEW.astronaut_id
        AND (ac.expiry_date IS NULL OR ac.expiry_date > CURRENT_DATE);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_certification_expiry
AFTER UPDATE OF expiry_date ON space_mission.astronaut_certifications
FOR EACH ROW
EXECUTE FUNCTION space_mission.update_certification_status();


CREATE OR REPLACE FUNCTION space_mission.log_mission_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        
        UPDATE space_mission.mission_status_history
        SET effective_to = CURRENT_TIMESTAMP
        WHERE mission_id = NEW.mission_id
        AND effective_to IS NULL;
        
        
        INSERT INTO space_mission.mission_status_history
        (mission_id, status, effective_from, modified_by_astronaut_id)
        VALUES
        (NEW.mission_id, NEW.status, CURRENT_TIMESTAMP, CURRENT_SETTING('space_mission.current_astronaut_id', TRUE)::INTEGER);
        
        
        IF NEW.status IN ('Completed', 'Cancelled') THEN
            UPDATE space_mission.objectives
            SET status = CASE 
                WHEN status = 'In Progress' AND NEW.status = 'Completed' THEN 'Completed'
                WHEN status = 'In Progress' AND NEW.status = 'Cancelled' THEN 'Cancelled'
                WHEN status = 'Not Started' AND NEW.status = 'Cancelled' THEN 'Cancelled'
                ELSE status
                END
            WHERE mission_id = NEW.mission_id
            AND status IN ('Not Started', 'In Progress');
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_mission_status_change
AFTER UPDATE OF status ON space_mission.missions
FOR EACH ROW
EXECUTE FUNCTION space_mission.log_mission_status_change();

ALTER TABLE space_mission.missions ENABLE ROW LEVEL SECURITY;


CREATE POLICY mission_view_all ON space_mission.missions
    FOR SELECT
    USING (TRUE);

CREATE POLICY mission_update_lead ON space_mission.missions
    FOR UPDATE
    USING (lead_astronaut_id = CURRENT_SETTING('space_mission.current_astronaut_id', TRUE)::INTEGER);

CREATE POLICY mission_update_client ON space_mission.missions
    FOR UPDATE
    USING (client_astronaut_id = CURRENT_SETTING('space_mission.current_astronaut_id', TRUE)::INTEGER);


ALTER TABLE space_mission.crew_assignments ENABLE ROW LEVEL SECURITY;

CREATE POLICY crew_assign_view_all ON space_mission.crew_assignments
    FOR SELECT
    USING (TRUE);

CREATE POLICY crew_assign_manage ON space_mission.crew_assignments
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM space_mission.missions m
            WHERE m.mission_id = crew_assignments.mission_id
            AND m.lead_astronaut_id = CURRENT_SETTING('space_mission.current_astronaut_id', TRUE)::INTEGER
        )
    );
