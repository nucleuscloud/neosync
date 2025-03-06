erDiagram

    %% --------------------------------------------------------
    %% 1. company / department / transaction / expense_report / expense / item
    %% --------------------------------------------------------

    company {
        bigint id PK
        text name
        text url
        int employee_count
        uuid uuid
    }
    department {
        bigint id PK
        text name
        text url
        bigint company_id FK
        bigint user_id
        uuid uuid
    }
    transaction {
        bigint id PK
        doublePrecision amount
        timestamp created
        timestamp updated
        bigint department_id FK
        date transactionDate
        text currency
        json settings
        text description
        text timezone
        uuid uuid
    }
    expense_report {
        bigint id PK
        text invoice_id
        date reportDate
        numeric_15_2 amount
        bigint department_source_id FK
        bigint department_destination_id FK
        timestamp created
        timestamp updated
        varchar_5 currency
        int transaction_type
        boolean paid
        bigint transaction_id FK
        numeric_15_2 adjustment_amount
    }
    expense {
        bigint id PK
        bigint report_id FK
    }
    item {
        bigint id PK
        bigint expense_id FK
    }

    company ||--|{ department : "department_company_id_fkey"
    department ||--|{ transaction : "transaction_department_id_fkey"
    department ||--|{ expense_report : "dept_source_fkey"
    department ||--|{ expense_report : "dept_destination_fkey"
    transaction ||--|{ expense_report : "expense_report_transaction_fkey"
    expense_report ||--|{ expense : "expense_report_d_fkey"
    expense ||--|{ item : "expense_id_fkey"

    %% --------------------------------------------------------
    %% 2. test_2_x / test_2_a / test_2_b / test_2_c / test_2_d / test_2_e
    %% --------------------------------------------------------
    test_2_x {
        bigint id PK
        text name
        timestamp created
    }
    test_2_b {
        bigint id PK
        text name
        timestamp created
    }
    test_2_a {
        bigint id PK
        bigint x_id FK
    }
    test_2_c {
        bigint id PK
        text name
        timestamp created
        bigint a_id FK
        bigint b_id FK
    }
    test_2_d {
        bigint id PK
        bigint c_id FK
    }
    test_2_e {
        bigint id PK
        bigint c_id FK
    }

    test_2_x ||--|{ test_2_a : "fk_x_in_a"
    test_2_a ||--|{ test_2_c : "fk_a_in_c"
    test_2_b ||--|{ test_2_c : "fk_b_in_c"
    test_2_c ||--|{ test_2_d : "fk_c_in_d"
    test_2_c ||--|{ test_2_e : "fk_c_in_e"

    %% --------------------------------------------------------
    %% 3. test_3_a / test_3_b / test_3_c / test_3_d / test_3_e
    %%    test_3_f / test_3_g / test_3_h / test_3_i
    %% --------------------------------------------------------

    test_3_a {
        bigint id PK
    }
    test_3_b {
        bigint id PK
        bigint a_id FK
    }
    test_3_c {
        bigint id PK
        bigint b_id FK
    }
    test_3_d {
        bigint id PK
        bigint c_id FK
    }
    test_3_e {
        bigint id PK
        bigint d_id FK
    }
    test_3_f {
        bigint id PK
    }
    test_3_g {
        bigint id PK
        bigint f_id FK
    }
    test_3_h {
        bigint id PK
        bigint g_id FK
    }
    test_3_i {
        bigint id PK
        bigint h_id FK
    }

    test_3_a ||--|{ test_3_b : "test3_a"
    test_3_b ||--|{ test_3_c : "test3_b"
    test_3_c ||--|{ test_3_d : "test3_c"
    test_3_d ||--|{ test_3_e : "test3_d"
    test_3_f ||--|{ test_3_g : "test3_f"
    test_3_g ||--|{ test_3_h : "test3_g"
    test_3_h ||--|{ test_3_i : "test3_h"


    %% --------------------------------------------------------
    %% 4. addresses / customers / orders / payments (circular dependency)
    %% --------------------------------------------------------
    addresses {
        bigint id PK
        bigint order_id FK
    }
    customers {
        bigint id PK
        bigint address_id FK
    }
    orders {
        bigint id PK
        bigint customer_id FK
    }
    payments {
        bigint id PK
        bigint customer_id FK
    }

    addresses ||--|{ customers : "fk_address_id"
    customers ||--|{ orders : "fk_customer_id"
    customers ||--|{ payments : "fk_customer_id"
    orders ||--|{ addresses : "fk_order_id" 

    %% --------------------------------------------------------
    %% 5. division / employees (composite PK) / projects (composite FK)
    %% --------------------------------------------------------
    division {
        bigint id PK
        varchar_100 division_name
        varchar_100 location
    }
    employees {
        bigint id PK
        bigint division_id PK
        varchar_50 first_name
        varchar_50 last_name
        varchar_100 email
    }
    projects {
        bigint id PK
        varchar_100 project_name
        date start_date
        date end_date
        bigint responsible_employee_id
        bigint responsible_division_id
    }

    division ||--|{ employees : "employees_division_fkey"
    employees ||--|{ projects : "fk_projects_employees"

    %% --------------------------------------------------------
    %% 6. bosses / minions (self-referencing)
    %% --------------------------------------------------------
    bosses {
        bigint id PK
        bigint manager_id FK
        bigint big_boss_id FK
    }
    minions {
        bigint id PK
        bigint boss_id FK
    }

    bosses ||--|{ bosses : "self_mgr_fk"
    bosses ||--|{ bosses : "self_big_boss_fk"
    bosses ||--|{ minions : "boss_id_fk"

    %% --------------------------------------------------------
    %% 7. users / initiatives / tasks / skills / user_skills / comments / attachments
    %% --------------------------------------------------------
    users {
        int user_id PK
        varchar_100 name
        varchar_100 email
        int manager_id FK
        int mentor_id FK
    }
    initiatives {
        int initiative_id PK
        varchar_100 name
        text description
        int lead_id FK
        int client_id FK
    }
    tasks {
        int task_id PK
        varchar_200 title
        text description
        varchar_50 status
        int initiative_id FK
        int assignee_id FK
        int reviewer_id FK
    }
    skills {
        int skill_id PK
        varchar_100 name
        varchar_100 category
    }
    user_skills {
        int user_skill_id PK
        int user_id FK
        int skill_id FK
        int proficiency_level
    }
    comments {
        int comment_id PK
        text content
        timestamp created_at
        int user_id FK
        int task_id FK
        int initiative_id FK
        int parent_comment_id FK
    }
    attachments {
        int attachment_id PK
        varchar_255 file_name
        varchar_255 file_path
        int uploaded_by FK
        int task_id FK
        int initiative_id FK
        int comment_id FK
    }

    users ||--|{ users : "fk_user_manager"
    users ||--|{ users : "fk_user_mentor"
    users ||--|{ initiatives : "fk_initiative_lead"
    users ||--|{ initiatives : "fk_initiative_client"
    initiatives ||--|{ tasks : "fk_task_initiative"
    users ||--|{ tasks : "fk_task_assignee"
    users ||--|{ tasks : "fk_task_reviewer"
    users ||--|{ user_skills : "fk_user_skill_user"
    skills ||--|{ user_skills : "fk_user_skill_skill"
    users ||--|{ comments : "fk_comment_user"
    tasks ||--|{ comments : "fk_comment_task"
    initiatives ||--|{ comments : "fk_comment_initiative"
    comments ||--|{ comments : "fk_comment_parent"
    users ||--|{ attachments : "fk_attachment_user"
    tasks ||--|{ attachments : "fk_attachment_task"
    initiatives ||--|{ attachments : "fk_attachment_initiative"
    comments ||--|{ attachments : "fk_attachment_comment"

    %% --------------------------------------------------------
    %% 8. network_types / networks / network_users
    %% --------------------------------------------------------
    network_types {
        int id PK
        varchar_10 name
    }
    networks {
        int id PK
        varchar_255 name
        varchar_45 address
        int network_type_id FK
    }
    network_users {
        int id PK
        varchar_50 username
        varchar_255 email
        varchar_255 password_hash
        varchar_50 first_name
        varchar_50 last_name
        int network_id FK
    }

    network_types ||--|{ networks : "fk_network_type"
    networks ||--|{ network_users : "fk_network_id"
