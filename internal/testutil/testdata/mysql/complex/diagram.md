erDiagram

    AGENCY ||--|{ ASTRONAUT : "employs"
    AGENCY ||--|{ SPACECRAFT : "owns"
    AGENCY ||--|{ MISSION : "primary agency"

    ASTRONAUT ||--|{ MISSION_CREW : "many-to-many link"

    SPACECRAFT }|--|| MISSION : "used by (spacecraft_id)"
    
    CELESTIAL_BODY ||--|{ CELESTIAL_BODY : "self reference (parent_body_id)"

    MISSION ||--|{ MISSION_CREW : "crew assignments"
    MISSION ||--|{ MISSION_LOG : "has logs"
    MISSION }|--|| SPACECRAFT : "FK: spacecraft_id"
    MISSION }|--|| CELESTIAL_BODY : "FK: destination_id"
    MISSION }|--|| LAUNCH_SITE : "FK: launch_site_id"
    MISSION }|--|| AGENCY : "FK: primary_agency_id"
    MISSION }|--|| ASTRONAUT : "commander"

    MISSION_CREW }|--|| MISSION : "FK: mission_id"
    MISSION_CREW }|--|| ASTRONAUT : "FK: astronaut_id"

    PROJECT_MISSION }|--|| RESEARCH_PROJECT : "FK: project_id"
    PROJECT_MISSION }|--|| MISSION : "FK: mission_id"

    MISSION_LOG }|--|| MISSION : "FK: mission_id"



    %% Table Definitions
    AGENCY {
        int id PK
        varchar name
        varchar country
        year founded_year
    }

    ASTRONAUT {
        int id PK
        varchar first_name
        varchar last_name
        date birth_date
        varchar nationality
        enum status
        int agency_id FK
        int first_mission_id FK
    }

    SPACECRAFT {
        int id PK
        varchar name
        enum type
        int capacity
        enum status
        int agency_id FK
        int last_mission_id FK
    }

    CELESTIAL_BODY {
        int id PK
        varchar name
        enum body_type
        double mass
        double radius
        int parent_body_id FK
    }

    LAUNCH_SITE {
        int id PK
        varchar name
        varchar location
        point location_coord
        varchar country
    }

    MISSION {
        int id PK
        varchar name
        varchar mission_code
        enum mission_type
        enum status
        date launch_date
        date return_date
        int spacecraft_id FK
        int destination_id FK
        int launch_site_id FK
        int primary_agency_id FK
        int commander_id FK
    }

    MISSION_CREW {
        int mission_id PK
        int astronaut_id PK
        enum role
    }

    RESEARCH_PROJECT {
        int id PK
        varchar title
        text description
        date start_date
        date end_date
        int lead_astronaut_id FK
    }

    PROJECT_MISSION {
        int project_id PK
        int mission_id PK
    }

    MISSION_LOG {
        int log_id PK
        int mission_id FK
        timestamp log_time
        varchar event
    }

%% ========== COSMIC_RESEARCH SCHEMA (new) ==========
OBSERVATORY {
        int id PK
        varchar name
        int agency_id FK
        int launch_site_id FK
        enum status
        point location_coord
    }

    TELESCOPE {
        int id PK
        int observatory_id FK
        varchar name
        enum telescope_type
        double mirror_diameter_m
        enum status
    }

    INSTRUMENT {
        int id PK
        varchar name
        enum instrument_type
        int telescope_id FK 
        int spacecraft_id FK
        enum status
    }

    OBSERVATION_SESSION {
        int id PK
        int telescope_id FK 
        int instrument_id FK
        int target_body_id FK
        int mission_id FK
        datetime start_time
        datetime end_time
        enum seeing_conditions
        text notes
    }

    DATA_SET {
        int id PK
        varchar name
        int mission_id FK
        int observation_id FK
        text data_description
        longblob data_blob
        date collected_on
    }

    RESEARCH_PAPER {
        int id PK
        varchar title
        text abstract
        date published_date
        varchar doi
        int project_id FK
        int observatory_id FK
    }

    PAPER_CITATION {
        int citing_paper_id PK, FK
        int cited_paper_id  PK, FK
        date citation_date
    }

    GRANT {
        int id PK
        varchar grant_number
        int agency_id FK
        decimal funding_amount
        date start_date
        date end_date
        enum status
    }

    GRANT_RESEARCH_PROJECT {
        int grant_id PK, FK
        int research_project_id PK, FK
        decimal allocated_amount
    }

    INSTRUMENT_USAGE {
        int id PK
        int instrument_id FK
        int telescope_id FK
        int spacecraft_id FK
        date start_date
        date end_date
        text usage_notes
    }

    %% Relationships among cosmic_research tables
    OBSERVATORY ||--|{ TELESCOPE : has
    TELESCOPE ||--|{ INSTRUMENT : can_host_instrument
    TELESCOPE ||--|{ OBSERVATION_SESSION : used_in
    INSTRUMENT ||--|{ OBSERVATION_SESSION : used_in
    OBSERVATION_SESSION ||--|{ DATA_SET : produces
    RESEARCH_PAPER }|..|| OBSERVATORY : from
    RESEARCH_PAPER }|..|| RESEARCH_PROJECT : pertains_to
    RESEARCH_PAPER ||--|{ PAPER_CITATION : citing
    RESEARCH_PAPER ||--|{ PAPER_CITATION : cited
    GRANT ||--|{ GRANT_RESEARCH_PROJECT : funds
    RESEARCH_PROJECT ||--|{ GRANT_RESEARCH_PROJECT : funded

    INSTRUMENT_USAGE ||--|| INSTRUMENT : usage_of
    INSTRUMENT_USAGE }|--|| TELESCOPE : usage_on_telescope
    INSTRUMENT_USAGE }|--|| SPACECRAFT : usage_on_spacecraft
