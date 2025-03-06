erDiagram

    %% -------------------------------------------------------
    %% 1) TABLES (Minimal column lists, PK/FK notations)
    %% -------------------------------------------------------

    ASTRONAUTS {
        int astronaut_id PK
        string name
        string email
        int manager_astronaut_id FK
        int mentor_astronaut_id FK
    }

    MISSIONS {
        int mission_id PK
        string name
        string description
        int lead_astronaut_id FK
        int client_astronaut_id FK
        string status
    }

    OBJECTIVES {
        int objective_id PK
        string title
        string description
        string status
        int mission_id FK
        int assignee_astronaut_id FK
        int reviewer_astronaut_id FK
    }

    CAPABILITIES {
        int capability_id PK
        string name
        string category
    }

    ASTRONAUT_CAPABILITIES {
        int astronaut_capability_id PK
        int astronaut_id FK
        int capability_id FK
        int proficiency_level
    }

    TRANSMISSIONS {
        int transmission_id PK
        string content
        string created_at
        int astronaut_id FK
        int objective_id FK
        int mission_id FK
        int parent_transmission_id FK
    }

    PAYLOADS {
        int payload_id PK
        string file_name
        string file_path
        int uploaded_by_astronaut_id FK
        int objective_id FK
        int mission_id FK
        int transmission_id FK
    }

    CREW_ASSIGNMENTS {
        int crew_assignment_id PK
        int astronaut_id FK
        int mission_id FK
        string role
    }

    MISSION_LOGS {
        int log_id PK
        string object_type
        int object_id
        string action
        string timestamp
    }

    CREWS {
        int crew_id PK
        string crew_name
        int lead_astronaut_id FK
        int parent_crew_id FK
    }

    CREW_MISSIONS {
        int crew_mission_id PK
        int crew_id FK
        int mission_id FK
        string notes
    }

    SUPPLIES {
        int supply_id PK
        int mission_id FK
        int bill_to_astronaut_id FK
        int owner_astronaut_id FK
        float total_amount
        string status
        string created_at
    }

    SUPPLY_ITEMS {
        int supply_item_id PK
        int supply_id FK
        int objective_id FK
        string description
        int quantity
        float unit_price
    }

    SPACECRAFT_CLASS {
        int class_id PK
        string class_name
        string description
    }

    SPACECRAFT {
        int spacecraft_id PK
        string name
        int class_id FK
        string launch_date
        string status
    }

    SPACECRAFT_MODULE {
        int module_id PK
        int spacecraft_id FK
        string module_name
        string purpose
        string installation_date
    }

    MODULE_COMPONENT {
        int component_id PK
        int module_id FK
        string component_name
        string manufacturer
        string serial_number
        string installation_date
    }

    EQUIPMENT {
        int equipment_id PK
        string name
        string category
        float weight_kg
    }

    MISSION_EQUIPMENT {
        int mission_id FK
        int equipment_id FK
        int quantity
        string assignment_date
        %% composite key: (mission_id, equipment_id)
    }

    EQUIPMENT_MAINTENANCE {
        int maintenance_id PK
        int mission_id FK
        int equipment_id FK
        string maintenance_date
        int performed_by_astronaut_id FK
        string description
    }

    TRAINING_COURSES {
        int course_id PK
        string course_name
        int duration_hours
        string description
    }

    COURSE_PREREQUISITES {
        int prerequisite_id PK
        int course_id FK
        int required_course_id FK
    }

    CERTIFICATIONS {
        int certification_id PK
        string name
        string issuing_authority
        int valid_years
    }

    ASTRONAUT_CERTIFICATIONS {
        int astronaut_id FK
        int certification_id FK
        string issue_date
        string expiry_date
        string certificate_number
        %% composite key: (astronaut_id, certification_id)
    }

    CERTIFICATION_REQUIREMENTS {
        int requirement_id PK
        int certification_id FK
        int required_certification_id FK
    }

    MISSION_LOGS_EXTENDED {
        int log_id PK
        int mission_id FK
        int astronaut_id FK
        string log_date
        string log_type
        string severity
        string message
    }

    COMMUNICATION_CHANNELS {
        int channel_id PK
        string name
        string frequency
        string encryption_level
    }

    MISSION_COMMUNICATIONS {
        int mission_id FK
        int channel_id FK
        int priority
        bool is_backup
        %% composite key: (mission_id, channel_id)
    }

    MESSAGE_LOGS {
        int log_id PK
        int mission_id FK
        int channel_id FK
        string timestamp
        int sender_astronaut_id FK
        string content
        bool is_encrypted
    }

    EVENTS {
        int event_id PK
        string event_time
        string severity
        string title
        string description
    }

    SYSTEM_EVENTS {
        %% inherits from EVENTS
        string system_name
        int component_id
        string error_code
        bool resolved
    }

    ASTRONAUT_EVENTS {
        %% inherits from EVENTS
        int astronaut_id FK
        string location
        string vital_signs
        int mission_id FK
    }

    MISSION_EVENTS {
        %% inherits from EVENTS
        int mission_id FK
        string milestone
        int success_rating
        string affected_objectives
    }

    TELEMETRY {
        int telemetry_id PK
        string timestamp
        int spacecraft_id FK
        string sensor_type
        float reading
        string unit
        string coordinates
        bool is_anomaly
        %% partitioned by range (timestamp)
    }

    COMMENTS {
        int comment_id PK
        string reference_type
        int reference_id
        int author_astronaut_id FK
        string content
        string created_at
        string metadata
    }

    TAGS {
        int tag_id PK
        string name
        string category
        int created_by_astronaut_id FK
    }

    TAGGABLES {
        int tag_id FK
        string taggable_type
        int taggable_id
        int created_by_astronaut_id FK
        %% composite key: (tag_id, taggable_type, taggable_id)
    }

    SCIENTIFIC_DATA_EXPERIMENTS {
        int experiment_id PK
        string name
        string description
        int lead_scientist_astronaut_id FK
        string start_date
        string end_date
        string status
        int mission_id FK
    }

    SCIENTIFIC_DATA_SAMPLES {
        int sample_id PK
        int experiment_id FK
        string collection_date
        string location
        int collected_by_astronaut_id FK
        string sample_type
        int mission_id FK
        int spacecraft_id FK
    }

    SCIENTIFIC_DATA_MEASUREMENTS {
        int measurement_id PK
        int sample_id FK
        string parameter
        float value
        string unit
        string measured_at
        int measured_by_astronaut_id FK
        bool is_verified
    }

    MISSION_EXPERIMENTS {
        int mission_id FK
        int experiment_id FK
        int priority
        int time_allocation_hours
        string resources_allocated
        string notes
        %% composite key: (mission_id, experiment_id)
    }

    MISSION_PARAMETERS {
        int parameter_id PK
        int mission_id FK
        string name
        string numeric_values
        string string_values
        string json_config
        string applicable_spacecraft
        string tags
    }

    SKILL_GROUPS {
        int group_id PK
        string name
        string description
        int parent_group_id FK
        string required_for_role
    }

    CAPABILITY_SKILL_GROUPS {
        int capability_id FK
        int group_id FK
        int importance_level
        %% composite key: (capability_id, group_id)
    }

    MISSION_REQUIRED_SKILL_GROUPS {
        int mission_id FK
        int group_id FK
        int minimum_proficiency
        string role
        bool is_critical
        %% composite key: (mission_id, group_id, role)
    }

    EQUIPMENT_COMPATIBILITY {
        int primary_equipment_id FK
        int compatible_equipment_id FK
        int compatibility_level
        string notes
        %% composite key: (primary_equipment_id, compatible_equipment_id)
    }

    ASTRONAUT_VITALS {
        int vital_id PK
        int astronaut_id FK
        int mission_id FK
        string location
        string vital_data
        string recorded_at
        bool is_anomalous
    }

    MISSION_STATUS_HISTORY {
        int history_id PK
        int mission_id FK
        string status
        string effective_from
        string effective_to
        int modified_by_astronaut_id FK
        string reason
    }

    EQUIPMENT_STATUS_HISTORY {
        int history_id PK
        int equipment_id FK
        int mission_id FK
        string status
        string effective_from
        string effective_to
        int condition_rating
        string notes
    }

    ASTRONAUT_ROLE_HISTORY {
        int history_id PK
        int astronaut_id FK
        string role
        int mission_id FK
        string effective_from
        string effective_to
        int supervisor_astronaut_id FK
        int performance_rating
    }

    %% -------------------------------------------------------
    %% 2) RELATIONSHIPS (Foreign Keys -> "A ||--|{ B : label")
    %% -------------------------------------------------------

    %% Self references in ASTRONAUTS
    ASTRONAUTS ||--|{ ASTRONAUTS : "manager_astronaut_id"
    ASTRONAUTS ||--|{ ASTRONAUTS : "mentor_astronaut_id"

    %% Missions -> Astronauts
    ASTRONAUTS ||--|{ MISSIONS : "lead_astronaut_id"
    ASTRONAUTS ||--|{ MISSIONS : "client_astronaut_id"

    %% Objectives
    MISSIONS ||--|{ OBJECTIVES : "mission_id"
    ASTRONAUTS ||--|{ OBJECTIVES : "assignee_astronaut_id"
    ASTRONAUTS ||--|{ OBJECTIVES : "reviewer_astronaut_id"

    %% Astronaut Capabilities
    ASTRONAUTS ||--|{ ASTRONAUT_CAPABILITIES : "astronaut_id"
    CAPABILITIES ||--|{ ASTRONAUT_CAPABILITIES : "capability_id"

    %% Transmissions
    ASTRONAUTS ||--|{ TRANSMISSIONS : "astronaut_id"
    OBJECTIVES ||--|{ TRANSMISSIONS : "objective_id"
    MISSIONS ||--|{ TRANSMISSIONS : "mission_id"
    TRANSMISSIONS ||--|{ TRANSMISSIONS : "parent_transmission_id"

    %% Payloads
    ASTRONAUTS ||--|{ PAYLOADS : "uploaded_by_astronaut_id"
    OBJECTIVES ||--|{ PAYLOADS : "objective_id"
    MISSIONS ||--|{ PAYLOADS : "mission_id"
    TRANSMISSIONS ||--|{ PAYLOADS : "transmission_id"

    %% Crew Assignments
    ASTRONAUTS ||--|{ CREW_ASSIGNMENTS : "astronaut_id"
    MISSIONS ||--|{ CREW_ASSIGNMENTS : "mission_id"

    %% Crews
    ASTRONAUTS ||--|{ CREWS : "lead_astronaut_id"
    CREWS ||--|{ CREWS : "parent_crew_id"

    %% Crew Missions
    CREWS ||--|{ CREW_MISSIONS : "crew_id"
    MISSIONS ||--|{ CREW_MISSIONS : "mission_id"

    %% Supplies & Supply Items
    MISSIONS ||--|{ SUPPLIES : "mission_id"
    ASTRONAUTS ||--|{ SUPPLIES : "bill_to_astronaut_id"
    ASTRONAUTS ||--|{ SUPPLIES : "owner_astronaut_id"
    SUPPLIES ||--|{ SUPPLY_ITEMS : "supply_id"
    OBJECTIVES ||--|{ SUPPLY_ITEMS : "objective_id"

    %% Spacecraft
    SPACECRAFT_CLASS ||--|{ SPACECRAFT : "class_id"
    SPACECRAFT ||--|{ SPACECRAFT_MODULE : "spacecraft_id"
    SPACECRAFT_MODULE ||--|{ MODULE_COMPONENT : "module_id"

    %% Equipment
    MISSIONS ||--|{ MISSION_EQUIPMENT : "mission_id"
    EQUIPMENT ||--|{ MISSION_EQUIPMENT : "equipment_id"
    MISSION_EQUIPMENT ||--|{ EQUIPMENT_MAINTENANCE : "mission_id,equipment_id"
    ASTRONAUTS ||--|{ EQUIPMENT_MAINTENANCE : "performed_by_astronaut_id"

    %% Training courses
    TRAINING_COURSES ||--|{ COURSE_PREREQUISITES : "course_id"
    TRAINING_COURSES ||--|{ COURSE_PREREQUISITES : "required_course_id"

    %% Certifications
    ASTRONAUTS ||--|{ ASTRONAUT_CERTIFICATIONS : "astronaut_id"
    CERTIFICATIONS ||--|{ ASTRONAUT_CERTIFICATIONS : "certification_id"
    CERTIFICATIONS ||--|{ CERTIFICATION_REQUIREMENTS : "certification_id"
    CERTIFICATIONS ||--|{ CERTIFICATION_REQUIREMENTS : "required_certification_id"

    %% Mission Logs Extended
    MISSIONS ||--|{ MISSION_LOGS_EXTENDED : "mission_id"
    ASTRONAUTS ||--o{ MISSION_LOGS_EXTENDED : "astronaut_id"

    %% Comms
    MISSIONS ||--|{ MISSION_COMMUNICATIONS : "mission_id"
    COMMUNICATION_CHANNELS ||--|{ MISSION_COMMUNICATIONS : "channel_id"
    MISSION_COMMUNICATIONS ||--o{ MESSAGE_LOGS : "mission_id,channel_id"
    ASTRONAUTS ||--o{ MESSAGE_LOGS : "sender_astronaut_id"

    %% Table Inheritance (events)
    EVENTS ||--o{ SYSTEM_EVENTS : "inherits"
    EVENTS ||--o{ ASTRONAUT_EVENTS : "inherits"
    EVENTS ||--o{ MISSION_EVENTS : "inherits"

    %% Telemetry partition
    SPACECRAFT ||--|{ TELEMETRY : "spacecraft_id"
    %% Partitions
    %% TELEMETRY ||--o{ TELEMETRY_2023 : "partition"
    %% etc. (not typically recognized by Mermaid, so omitted)

    %% Polymorphic
    ASTRONAUTS ||--o{ COMMENTS : "author_astronaut_id"
    TAGS ||--|{ TAGGABLES : "tag_id"
    ASTRONAUTS ||--o{ TAGGABLES : "created_by_astronaut_id"

    %% Multi-schema references
    ASTRONAUTS ||--|{ SCIENTIFIC_DATA_EXPERIMENTS : "lead_scientist_astronaut_id"
    MISSIONS ||--|{ SCIENTIFIC_DATA_EXPERIMENTS : "mission_id"
    SCIENTIFIC_DATA_EXPERIMENTS ||--|{ SCIENTIFIC_DATA_SAMPLES : "experiment_id"
    ASTRONAUTS ||--o{ SCIENTIFIC_DATA_SAMPLES : "collected_by_astronaut_id"
    MISSIONS ||--o{ SCIENTIFIC_DATA_SAMPLES : "mission_id"
    SPACECRAFT ||--o{ SCIENTIFIC_DATA_SAMPLES : "spacecraft_id"
    SCIENTIFIC_DATA_SAMPLES ||--|{ SCIENTIFIC_DATA_MEASUREMENTS : "sample_id"
    ASTRONAUTS ||--o{ SCIENTIFIC_DATA_MEASUREMENTS : "measured_by_astronaut_id"
    MISSIONS ||--|{ MISSION_EXPERIMENTS : "mission_id"
    SCIENTIFIC_DATA_EXPERIMENTS ||--|{ MISSION_EXPERIMENTS : "experiment_id"

    %% Mission Parameters
    MISSIONS ||--|{ MISSION_PARAMETERS : "mission_id"

    %% Skill Groups
    SKILL_GROUPS ||--|{ SKILL_GROUPS : "parent_group_id"
    CAPABILITIES ||--|{ CAPABILITY_SKILL_GROUPS : "capability_id"
    SKILL_GROUPS ||--|{ CAPABILITY_SKILL_GROUPS : "group_id"
    MISSIONS ||--|{ MISSION_REQUIRED_SKILL_GROUPS : "mission_id"
    SKILL_GROUPS ||--|{ MISSION_REQUIRED_SKILL_GROUPS : "group_id"

    %% Equipment Compatibility (self-ref)
    EQUIPMENT ||--|{ EQUIPMENT_COMPATIBILITY : "primary_equipment_id"
    EQUIPMENT ||--|{ EQUIPMENT_COMPATIBILITY : "compatible_equipment_id"

    %% Astronaut Vitals
    ASTRONAUTS ||--|{ ASTRONAUT_VITALS : "astronaut_id"
    MISSIONS ||--|{ ASTRONAUT_VITALS : "mission_id"

    %% Temporal History
    MISSIONS ||--|{ MISSION_STATUS_HISTORY : "mission_id"
    ASTRONAUTS ||--o{ MISSION_STATUS_HISTORY : "modified_by_astronaut_id"
    EQUIPMENT ||--|{ EQUIPMENT_STATUS_HISTORY : "equipment_id"
    MISSIONS ||--o{ EQUIPMENT_STATUS_HISTORY : "mission_id"
    ASTRONAUTS ||--|{ ASTRONAUT_ROLE_HISTORY : "astronaut_id"
    MISSIONS ||--o{ ASTRONAUT_ROLE_HISTORY : "mission_id"
    ASTRONAUTS ||--o{ ASTRONAUT_ROLE_HISTORY : "supervisor_astronaut_id"
