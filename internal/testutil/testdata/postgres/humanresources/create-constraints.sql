
ALTER TABLE countries
    ADD CONSTRAINT fk_countries_region_id
    FOREIGN KEY (region_id) REFERENCES regions (region_id); 

ALTER TABLE locations
    ADD CONSTRAINT fk_locations_country_id
    FOREIGN KEY (country_id) REFERENCES countries (country_id); 

ALTER TABLE departments
    ADD CONSTRAINT fk_departments_location_id
    FOREIGN KEY (location_id) REFERENCES locations (location_id); 


ALTER TABLE employees
    ADD CONSTRAINT fk_employees_job_id
    FOREIGN KEY (job_id) REFERENCES jobs (job_id); 

ALTER TABLE employees
    ADD CONSTRAINT fk_employees_department_id
    FOREIGN KEY (department_id) REFERENCES departments (department_id); 

ALTER TABLE employees
    ADD CONSTRAINT fk_employees_manager_id
    FOREIGN KEY (manager_id) REFERENCES employees (employee_id); 

ALTER TABLE dependents
    ADD CONSTRAINT fk_dependents_employee_id
    FOREIGN KEY (employee_id) REFERENCES employees (employee_id) 
