INSERT INTO astronauts (name, email, manager_astronaut_id, mentor_astronaut_id) VALUES
('John Doe', 'john.doe@example.com', NULL, NULL),
('Jane Smith', 'jane.smith@example.com', 1, NULL),
('Bob Johnson', 'bob.johnson@example.com', 1, 2),
('Alice Williams', 'alice.williams@example.com', 2, 1),
('Charlie Brown', 'charlie.brown@example.com', 2, 3),
('Diana Prince', 'diana.prince@example.com', 3, 4),
('Ethan Hunt', 'ethan.hunt@example.com', 3, 1),
('Fiona Gallagher', 'fiona.gallagher@example.com', 4, 2),
('George Lucas', 'george.lucas@example.com', 4, 5),
('Hannah Montana', 'hannah.montana@example.com', 5, 3);

INSERT INTO missions (name, description, status, lead_astronaut_id, client_astronaut_id) VALUES
('Website Redesign', 'Overhaul the company website', 'In Progress', 1, 2),
('Mobile App Development', 'Create a new mobile app', 'Active', 2, 3),
('Data Migration', 'Migrate data to new system', 'In Progress', 3, 4),
('AI Integration', 'Implement AI in current products', 'Research', 4, 5),
('Cloud Migration', 'Move infrastructure to the cloud', 'In Progress', 5, 6),
('Security Audit', 'Perform a comprehensive security audit', 'Scheduled', 6, 7),
('Performance Optimization', 'Optimize system performance', 'In Progress', 7, 8),
('Customer Portal', 'Develop a new customer portal', 'Active', 8, 9),
('Blockchain Implementation', 'Implement blockchain technology', 'Research', 9, 10),
('IoT Platform', 'Develop an IoT management platform', 'Active', 10, 1);

INSERT INTO objectives (title, description, status, mission_id, assignee_astronaut_id, reviewer_astronaut_id) VALUES
('Design mockups', 'Create initial design mockups', 'In Progress', 1, 3, 1),
('Develop login system', 'Implement secure login system', 'Not Started', 2, 4, 2),
('Data mapping', 'Map data fields between systems', 'Completed', 3, 5, 3),
('Train AI model', 'Train and test initial AI model', 'In Progress', 4, 6, 4),
('Setup cloud environment', 'Initialize cloud infrastructure', 'In Progress', 5, 7, 5),
('Vulnerability assessment', 'Identify system vulnerabilities', 'Not Started', 6, 8, 6),
('Code profiling', 'Profile code for performance bottlenecks', 'In Progress', 7, 9, 7),
('Design user interface', 'Design intuitive user interface', 'Completed', 8, 10, 8),
('Smart contract development', 'Develop initial smart contracts', 'In Progress', 9, 1, 9),
('Sensor integration', 'Integrate IoT sensors with platform', 'Not Started', 10, 2, 10);

INSERT INTO capabilities (name, category) VALUES
('JavaScript', 'Programming'),
('Python', 'Programming'),
('SQL', 'Database'),
('Project Management', 'Management'),
('UI/UX Design', 'Design'),
('Machine Learning', 'Data Science'),
('Network Security', 'Security'),
('Cloud Architecture', 'Infrastructure'),
('Blockchain', 'Technology'),
('IoT', 'Technology');

INSERT INTO astronaut_capabilities (astronaut_id, capability_id, proficiency_level) VALUES
(1, 1, 5),
(2, 2, 4),
(3, 3, 5),
(4, 4, 4),
(5, 5, 3),
(6, 6, 5),
(7, 7, 4),
(8, 8, 3),
(9, 9, 4),
(10, 10, 5);

INSERT INTO transmissions (
    transmission_id, content, astronaut_id, objective_id, mission_id, parent_transmission_id
) VALUES
(1, 'Great progress on the mockups!', 1, 1, 1, NULL),
(2, 'Thanks! I’ve incorporated the feedback.', 3, 1, 1, 1),
(3, 'We need to use OAuth for the login system', 2, 2, 2, NULL),
(4, 'Agreed. I’ll update the design docs.', 4, 2, 2, 3),
(5, 'Data mapping completed, ready for review', 3, 3, 3, NULL),
(6, 'I’ll start the review today.', 5, 3, 3, 5),
(7, 'AI model showing promising results', 4, 4, 4, NULL),
(8, 'That’s great news! Demo soon?', 6, 4, 4, 7),
(9, 'Cloud environment set up', 5, 5, 5, NULL),
(10, 'Let’s begin the migration.', 7, 5, 5, 9),
(11, 'Found critical vulnerabilities', 6, 6, 6, NULL),
(12, 'Can you prioritize them?', 8, 6, 6, 11),
(13, 'Performance bottleneck identified', 7, 7, 7, NULL),
(14, 'Plan to address it?', 9, 7, 7, 13),
(15, 'UI designs are approved', 8, 8, 8, NULL),
(16, 'Great! Next step: dev', 10, 8, 8, 15),
(17, 'Smart contracts pass tests', 9, 9, 9, NULL),
(18, 'Schedule a security audit next.', 1, 9, 9, 17),
(19, 'Sensor compatibility issues', 10, 10, 10, NULL),
(20, 'I can help troubleshoot.', 2, 10, 10, 19);

INSERT INTO payloads (
    file_name, file_path, uploaded_by_astronaut_id, objective_id, mission_id, transmission_id
) VALUES
('mockup_v1.png', '/files/mockups/mockup_v1.png', 3, 1, 1, 2),
('login_flow.pdf', '/files/docs/login_flow.pdf', 4, 2, 2, 4),
('data_mapping.xlsx', '/files/data/data_mapping.xlsx', 5, 3, 3, 5),
('ai_model_results.ipynb', '/files/notebooks/ai_model_results.ipynb', 6, 4, 4, 7),
('cloud_architecture.jpg', '/files/diagrams/cloud_architecture.jpg', 7, 5, 5, 9),
('security_report.pdf', '/files/reports/security_report.pdf', 8, 6, 6, 11),
('performance_analysis.html', '/files/reports/performance_analysis.html', 9, 7, 7, 13),
('ui_designs.sketch', '/files/designs/ui_designs.sketch', 10, 8, 8, 15),
('smart_contracts.sol', '/files/blockchain/smart_contracts.sol', 1, 9, 9, 17),
('sensor_specs.pdf', '/files/iot/sensor_specs.pdf', 2, 10, 10, 19);

INSERT INTO crew_assignments (astronaut_id, mission_id, role) VALUES
(3, 1, 'Designer'),
(4, 2, 'Developer'),
(5, 3, 'Engineer'),
(6, 4, 'Researcher'),
(7, 5, 'Security Specialist'),
(8, 6, 'Systems Analyst'),
(9, 7, 'Database Administrator'),
(10, 8, 'Project Manager'),
(1, 9, 'Technical Lead'),
(2, 10, 'Quality Assurance');

INSERT INTO mission_logs (object_type, object_id, action) VALUES
('astronauts', 1, 'Logged in'),
('astronauts', 999, 'Non-existent user action'),
('objectives', 3, 'Status changed'),
('capabilities', 15, 'Unreal skill updated'),
('missions', 2, 'Mission paused'),
('transmissions', 999, 'Invalid transmission reference');

INSERT INTO crews (crew_id, crew_name, lead_astronaut_id, parent_crew_id) VALUES
(1, 'Core Team', 1, NULL);


INSERT INTO crews (crew_id, crew_name, lead_astronaut_id, parent_crew_id) VALUES
(2, 'UX Subteam', 3, 1),
(3, 'Dev Subteam', 4, 1),
(4, 'Ops Subteam', 5, 1);

INSERT INTO crews (crew_id, crew_name, lead_astronaut_id, parent_crew_id) VALUES
(5, 'Ops Child Subteam', 6, 4);

INSERT INTO crew_missions (crew_id, mission_id, notes) VALUES
(1, 1, 'Core Team on Website Redesign'),
(2, 2, 'UX Subteam on Mobile App'),
(3, 9, 'Dev Subteam on Blockchain Impl.'),
(4, 3, 'Ops Subteam on Data Migration');

INSERT INTO supplies (mission_id, bill_to_astronaut_id, owner_astronaut_id, total_amount, status)
VALUES
(1, 2, 1, 1200.00, 'Open'),    
(2, 4, 2, 2000.00, 'Open'),
(3, 5, 3, 1850.00, 'Pending'),
(4, 6, 1, 9999.99, 'Open'),    
(5, 7, 3, 500.00, 'Open'),    
(6, 8, 4, 1500.00, 'Pending'),
(7, 9, 5, 750.00, 'Approved'),
(8, 10, 6, 2250.00, 'Open'),
(9, 1, 7, 3000.00, 'Pending'),
(10, 2, 8, 1750.00, 'Approved');

INSERT INTO supply_items (supply_id, objective_id, description, quantity, unit_price) VALUES
(8 , 1, 'Mockup design line item', 2, 150),
(6, NULL, 'Project management fee', 1, 200),
(2, 2, 'Login system dev cost', 1, 500),
(3, 3, 'Data mapping cost', 2, 250),
(7, 4, 'AI model training costs', 1, 999.99); 

INSERT INTO spacecraft_class (class_name, description) VALUES
('Explorer', 'Long-range exploration spacecraft'),
('Shuttle', 'Short-range transport spacecraft'),
('Station', 'Orbital research facilities');

INSERT INTO spacecraft (name, class_id, launch_date, status) VALUES
('Discovery', 1, '2030-01-15', 'Active'),
('Voyager', 1, '2032-03-22', 'Active'),
('Atlantis', 2, '2029-11-05', 'Maintenance');

INSERT INTO spacecraft_module (spacecraft_id, module_name, purpose, installation_date) VALUES
(1, 'Command Module', 'Navigation and control', '2029-12-10'),
(1, 'Research Lab', 'Scientific experiments', '2029-12-12'),
(2, 'Habitat Module', 'Crew living quarters', '2032-02-15');

INSERT INTO module_component (module_id, component_name, manufacturer, serial_number, installation_date) VALUES
(1, 'Navigation Computer', 'SpaceTech Inc', 'NT-12345', '2029-12-09'),
(1, 'Life Support System', 'VitalAir Systems', 'LS-98765', '2029-12-10'),
(2, 'Spectrometer', 'ScienceGear', 'SG-24680', '2029-12-12');

INSERT INTO equipment (name, category, weight_kg) VALUES
('Oxygen Tank', 'Life Support', 25.5),
('Solar Panel', 'Power', 35.2),
('Rover', 'Transportation', 150.0);

INSERT INTO mission_equipment (mission_id, equipment_id, quantity, assignment_date) VALUES
(1, 1, 5, '2023-01-10'),
(1, 2, 8, '2023-01-12'),
(2, 1, 3, '2023-02-05');

INSERT INTO equipment_maintenance (mission_id, equipment_id, maintenance_date, performed_by_astronaut_id, description) VALUES
(1, 1, '2023-03-15', 3, 'Routine inspection and pressure check'),
(1, 2, '2023-03-20', 4, 'Clean solar cells and check connections'),
(2, 1, '2023-04-10', 5, 'Replace pressure valve');

INSERT INTO training_courses (course_name, duration_hours, description) VALUES
('Basic Astronaut Training', 120, 'Fundamental training for all astronauts'),
('Advanced EVA Procedures', 80, 'Extravehicular activity advanced techniques'),
('Space Station Operations', 100, 'Operating procedures for space station systems');

INSERT INTO course_prerequisites (course_id, required_course_id) VALUES
(2, 1), 
(3, 1), 
(3, 2); 

INSERT INTO certifications (name, issuing_authority, valid_years) VALUES
('Basic Astronaut', 'International Space Agency', 5),
('EVA Specialist', 'International Space Agency', 3),
('Medical Officer', 'Space Medicine Board', 4);

INSERT INTO astronaut_certifications (astronaut_id, certification_id, issue_date, expiry_date, certificate_number) VALUES
(1, 1, '2022-05-10', '2027-05-10', 'BA-12345'),
(1, 2, '2022-07-15', '2025-07-15', 'EVA-54321'),
(2, 1, '2022-06-20', '2027-06-20', 'BA-23456'),
(3, 3, '2022-08-05', '2026-08-05', 'MO-87654');

INSERT INTO certification_requirements (certification_id, required_certification_id) VALUES
(2, 1), 
(3, 1), 
(3, 2); 

INSERT INTO mission_logs_extended (mission_id, astronaut_id, log_date, log_type, severity, message) VALUES
(1, 1, '2023-01-15 08:30:00', 'System Check', 'Info', 'All systems nominal'),
(1, NULL, '2023-01-15 12:45:00', 'Automated Alert', 'Warning', 'Oxygen level fluctuation detected'),
(2, 3, '2023-02-20 10:15:00', 'Activity Report', 'Info', 'Completed experiment setup');

INSERT INTO communication_channels (name, frequency, encryption_level) VALUES
('Primary Comms', '2.4 GHz', 'High'),
('Emergency Channel', '121.5 MHz', 'Medium'),
('Science Data Link', '5.8 GHz', 'High');

INSERT INTO mission_communications (mission_id, channel_id, priority, is_backup) VALUES
(1, 1, 1, FALSE),
(1, 2, 2, TRUE),
(2, 1, 1, FALSE);

INSERT INTO message_logs (mission_id, channel_id, timestamp, sender_astronaut_id, content, is_encrypted) VALUES
(1, 1, '2023-01-15 09:30:00', 1, 'Mission day 1 status report: all nominal', TRUE),
(1, NULL, '2023-01-15 14:20:00', NULL, 'AUTOMATED: Scheduled data transmission complete', FALSE),
(2, 1, '2023-02-22 11:45:00', 3, 'Experiment results ready for download', TRUE);

INSERT INTO space_mission.events (event_id, event_time, severity, title, description)
VALUES 
(1, '2023-05-15 08:22:15', 'Warning', 'Life Support Pressure Fluctuation', 'Minor pressure fluctuation detected in module A3'),
(2, '2023-05-16 14:37:22', 'Critical', 'Navigation System Failure', 'Primary navigation system offline'),
(3, '2023-05-17 22:14:05', 'Info', 'Routine System Check', 'All systems nominal after scheduled maintenance'),
(4, '2023-05-15 09:45:30', 'Info', 'EVA Started', 'Astronaut began scheduled EVA'),
(5, '2023-05-16 11:22:18', 'Warning', 'Elevated Heart Rate', 'Astronaut showing signs of exertion during repair procedure'),
(6, '2023-05-17 14:10:45', 'Emergency', 'Medical Incident', 'Astronaut reporting severe headache and disorientation'),
(7, '2023-05-15 00:00:00', 'Info', 'Mission Launch', 'Successful launch of mission'),
(8, '2023-05-16 10:30:00', 'Info', 'Docking Complete', 'Successfully docked with space station'),
(9, '2023-05-17 16:45:22', 'Warning', 'Experiment Complication', 'Unexpected results in primary experiment');

-- INSERT INTO system_events (
--     event_time, severity, title, description, 
--     system_name, component_id, error_code, resolved
-- ) VALUES
-- ('2023-05-15 08:22:15', 'Warning', 'Life Support Pressure Fluctuation', 'Minor pressure fluctuation detected in module A3', 
--  'Life Support', 235, 'LS-P-054', true),
-- ('2023-05-16 14:37:22', 'Critical', 'Navigation System Failure', 'Primary navigation system offline', 
--  'Navigation', 112, 'NAV-023', false),
-- ('2023-05-17 22:14:05', 'Info', 'Routine System Check', 'All systems nominal after scheduled maintenance', 
--  'Central Computer', 001, NULL, true);

-- INSERT INTO astronaut_events (
--     event_time, severity, title, description, 
--     astronaut_id, location, vital_signs, mission_id
-- ) VALUES
-- ('2023-05-15 09:45:30', 'Info', 'EVA Started', 'Astronaut began scheduled EVA', 
--  3, 'Exterior Module B', '{"heart_rate": 85, "blood_pressure": "120/80", "oxygen_saturation": 98.5}', 1),
-- ('2023-05-16 11:22:18', 'Warning', 'Elevated Heart Rate', 'Astronaut showing signs of exertion during repair procedure', 
--  4, 'Engine Room', '{"heart_rate": 115, "blood_pressure": "135/90", "oxygen_saturation": 97.2}', 2),
-- ('2023-05-17 14:10:45', 'Emergency', 'Medical Incident', 'Astronaut reporting severe headache and disorientation', 
--  5, 'Living Quarters', '{"heart_rate": 95, "blood_pressure": "145/95", "temperature": 38.5, "oxygen_saturation": 96.0}', 3);

-- INSERT INTO mission_events (
--     event_time, severity, title, description, 
--     mission_id, milestone, success_rating, affected_objectives
-- ) VALUES
-- ('2023-05-15 00:00:00', 'Info', 'Mission Launch', 'Successful launch of mission', 
--  1, 'Launch', 9, '{1,2,3}'),
-- ('2023-05-16 10:30:00', 'Info', 'Docking Complete', 'Successfully docked with space station', 
--  2, 'Docking', 8, '{4,5}'),
-- ('2023-05-17 16:45:22', 'Warning', 'Experiment Complication', 'Unexpected results in primary experiment', 
--  3, 'Research', 6, '{7,8}');

-- INSERT INTO telemetry (timestamp, spacecraft_id, sensor_type, reading, unit, coordinates, is_anomaly) VALUES
-- ('2023-05-15 00:00:01', 1, 'Temperature', 22.5, 'Celsius', '(45.2, 22.1)', false),
-- ('2023-05-15 00:01:01', 1, 'Pressure', 101.3, 'kPa', '(45.2, 22.1)', false),
-- ('2023-05-15 00:02:01', 1, 'Radiation', 0.05, 'mSv', '(45.2, 22.1)', false),
-- ('2023-05-16 12:30:01', 2, 'Temperature', 23.1, 'Celsius', '(46.3, 23.5)', false),
-- ('2023-05-16 12:31:01', 2, 'Pressure', 98.7, 'kPa', '(46.3, 23.5)', true),
-- ('2023-05-16 12:32:01', 2, 'Radiation', 0.12, 'mSv', '(46.3, 23.5)', false);

INSERT INTO comments (reference_type, reference_id, author_astronaut_id, content, metadata) VALUES
('missions', 1, 1, 'Launch conditions look optimal', '{"visibility": "public", "priority": "normal"}'),
('objectives', 3, 2, 'Data mapping is more complex than anticipated', '{"visibility": "team", "priority": "high"}'),
('astronauts', 5, 3, 'Charlie has shown exceptional problem-solving skills', '{"visibility": "management", "priority": "normal"}'),
('spacecraft', 1, 4, 'Discovery needs thruster recalibration before next mission', '{"visibility": "technical", "priority": "high"}');

INSERT INTO tags (name, category, created_by_astronaut_id) VALUES
('Urgent', 'Priority', 1),
('Technical', 'Category', 2),
('Research', 'Category', 3),
('Maintenance', 'Type', 4);

INSERT INTO taggables (tag_id, taggable_type, taggable_id, created_by_astronaut_id) VALUES
(1, 'objectives', 2, 1),
(2, 'objectives', 2, 1),
(3, 'missions', 4, 3),
(4, 'spacecraft', 1, 2);

INSERT INTO scientific_data.experiments (name, description, lead_scientist_astronaut_id, start_date, end_date, status, mission_id) VALUES
('Microgravity Crystal Growth', 'Study of protein crystal formation in microgravity', 6, '2023-04-01', '2023-06-30', 'In Progress', 4),
('Radiation Exposure Effects', 'Monitoring radiation effects on organic material', 1, '2023-03-15', '2023-07-15', 'In Progress', 5),
('Zero-G Fluid Dynamics', 'Study of fluid behavior in zero gravity', 4, '2023-05-01', '2023-08-30', 'Planned', 3),
('Microgravity Crystal Growth', 'Study of protein crystal formation in microgravity', 3, '2023-04-01', '2023-06-30', 'In Progress', 4),
('Radiation Exposure Effects', 'Monitoring radiation effects on organic material', 7, '2023-03-15', '2023-07-15', 'In Progress', 5),
('Zero-G Fluid Dynamics', 'Study of fluid behavior in zero gravity', 8, '2023-05-01', '2023-08-30', 'Planned', 6),
('Microgravity Crystal Growth', 'Study of protein crystal formation in microgravity', 6, '2023-04-01', '2023-06-30', 'In Progress', 4),
('Radiation Exposure Effects', 'Monitoring radiation effects on organic material', 7, '2023-03-15', '2023-07-15', 'In Progress', 5),
('Zero-G Fluid Dynamics', 'Study of fluid behavior in zero gravity', 8, '2023-05-01', '2023-08-30', 'Planned', 6);

INSERT INTO scientific_data.samples (experiment_id, collection_date, location, collected_by_astronaut_id, sample_type, storage_conditions, mission_id, spacecraft_id) VALUES
(1, '2023-04-15 09:30:00', 'Lab Module A', 6, 'Protein Crystal', 'Temperature: -20C, Pressure: 1atm', 4, 1),
(5, '2023-04-22 11:15:00', 'Lab Module A', 9, 'Protein Crystal', 'Temperature: -20C, Pressure: 1atm', 4, 1),
(9, '2023-04-10 14:45:00', 'External Platform', 7, 'Radiation Dosimeter', 'Ambient', 5, 2),
(6, '2023-04-15 09:30:00', 'Lab Module A', 6, 'Protein Crystal', 'Temperature: -20C, Pressure: 1atm', 4, 1),
(8, '2023-04-22 11:15:00', 'Lab Module A', 9, 'Protein Crystal', 'Temperature: -20C, Pressure: 1atm', 4, 1),
(2, '2023-04-10 14:45:00', 'External Platform', 2, 'Radiation Dosimeter', 'Ambient', 5, 2),
(3, '2023-04-15 09:30:00', 'Lab Module A', 4, 'Protein Crystal', 'Temperature: -20C, Pressure: 1atm', 4, 1),
(4, '2023-04-22 11:15:00', 'Lab Module A', 3, 'Protein Crystal', 'Temperature: -20C, Pressure: 1atm', 4, 1),
(2, '2023-04-10 14:45:00', 'External Platform', 4, 'Radiation Dosimeter', 'Ambient', 5, 2);

INSERT INTO scientific_data.measurements (sample_id, parameter, value, unit, measured_at, measured_by_astronaut_id, instrument, confidence_level, is_verified) VALUES
(1, 'Crystal Size', 2.45, 'mm', '2023-04-20 10:30:00', 6, 'Digital Microscope', 95.5, true),
(8, 'Density', 1.23, 'g/cm³', '2023-04-20 10:45:00', 2, 'Mass Spectrometer', 92.0, true),
(2, 'Crystal Size', 3.12, 'mm', '2023-04-25 09:15:00', 9, 'Digital Microscope', 94.0, false),
(1, 'Crystal Size', 2.45, 'mm', '2023-04-20 10:30:00', 6, 'Digital Microscope', 95.5, true),
(3, 'Density', 1.23, 'g/cm³', '2023-04-20 10:45:00', 6, 'Mass Spectrometer', 92.0, true),
(2, 'Crystal Size', 3.12, 'mm', '2023-04-25 09:15:00', 9, 'Digital Microscope', 94.0, false),
(9, 'Crystal Size', 2.45, 'mm', '2023-04-20 10:30:00', 1, 'Digital Microscope', 95.5, true),
(7, 'Density', 1.23, 'g/cm³', '2023-04-20 10:45:00', 4, 'Mass Spectrometer', 92.0, true),
(6, 'Crystal Size', 3.12, 'mm', '2023-04-25 09:15:00', 3, 'Digital Microscope', 94.0, false);

INSERT INTO space_mission.mission_experiments (mission_id, experiment_id, priority, time_allocation_hours, resources_allocated, notes) VALUES
(4, 1, 1, 120, '{"equipment": ["Microscope", "Centrifuge"], "supplies": ["Protein Samples", "Buffer Solution"]}', 'Primary mission objective'),
(5, 2, 2, 80, '{"equipment": ["Radiation Detector", "Shielding Materials"], "supplies": ["Organic Samples"]}', 'Secondary objective'),
(6, 3, 1, 100, '{"equipment": ["Fluid Containers", "High-Speed Camera"], "supplies": ["Test Fluids"]}', 'Critical for future missions'),
(4, 3, 1, 120, '{"equipment": ["Microscope", "Centrifuge"], "supplies": ["Protein Samples", "Buffer Solution"]}', 'Primary mission objective'),
(5, 1, 2, 80, '{"equipment": ["Radiation Detector", "Shielding Materials"], "supplies": ["Organic Samples"]}', 'Secondary objective'),
(6, 4, 1, 100, '{"equipment": ["Fluid Containers", "High-Speed Camera"], "supplies": ["Test Fluids"]}', 'Critical for future missions'),
(4, 2, 1, 120, '{"equipment": ["Microscope", "Centrifuge"], "supplies": ["Protein Samples", "Buffer Solution"]}', 'Primary mission objective'),
(5, 4, 2, 80, '{"equipment": ["Radiation Detector", "Shielding Materials"], "supplies": ["Organic Samples"]}', 'Secondary objective'),
(6, 1, 1, 100, '{"equipment": ["Fluid Containers", "High-Speed Camera"], "supplies": ["Test Fluids"]}', 'Critical for future missions');


INSERT INTO space_mission.mission_parameters (
    mission_id, name, numeric_values, string_values, json_config, applicable_spacecraft, tags
) VALUES
(1, 'Launch Parameters', '{32.5, 40.2, 15.7}', '{Stage 1, Stage 2, Orbit}', 
  '{"thrust_levels": {"stage1": 85, "stage2": 65}, "fuel_mixture": "standard", "trajectory": "optimized"}',
  '{1, 2}', '{launch, propulsion, trajectory}'),
(2, 'Communication Protocols', '{2.4, 5.8, 7.1}', '{Primary, Backup, Emergency}', 
  '{"encryption": "high", "bandwidth": 2.5, "channels": ["alpha", "beta", "gamma"]}',
  '{1, 2, 3}', '{communication, security, protocol}'),
(3, 'Environmental Controls', '{22.5, 40, 60}', '{Temperature, Humidity, Oxygen}', 
  '{"temperature_range": {"min": 20, "max": 25}, "humidity_target": 40, "oxygen_level": 21}',
  '{2, 3}', '{climate, life-support, environment}');

INSERT INTO space_mission.skill_groups (group_id, name, description, parent_group_id, required_for_role) VALUES
(1, 'Technical Skills', 'Core technical capabilities', NULL, '{Engineer, Technician}'),
(2, 'Piloting Skills', 'Spacecraft operation capabilities', NULL, '{Pilot, Commander}'),
(3, 'Scientific Skills', 'Research and analysis capabilities', NULL, '{Scientist, Researcher}'),
(4, 'Propulsion Systems', 'Engine and propulsion knowledge', 1, '{Propulsion Engineer}'),
(5, 'Navigation Systems', 'Navigation and guidance knowledge', 2, '{Navigator, Pilot}'),
(6, 'Biology', 'Biological science expertise', 3, '{Biologist, Medical Officer}');

INSERT INTO space_mission.capability_skill_groups (capability_id, group_id, importance_level) VALUES
(1, 1, 5),  
(2, 1, 4),  
(2, 3, 3),  
(3, 1, 5),  
(6, 3, 5),  
(7, 1, 4),  
(8, 4, 5),  
(10, 5, 4); 

INSERT INTO space_mission.mission_required_skill_groups (mission_id, group_id, minimum_proficiency, role, is_critical) VALUES
(1, 1, 4, 'Lead Engineer', true),
(1, 3, 3, 'Science Officer', false),
(2, 2, 5, 'Chief Pilot', true),
(2, 5, 4, 'Navigation Officer', true),
(3, 1, 3, 'Systems Engineer', false),
(4, 3, 5, 'Lead Scientist', true),
(4, 6, 4, 'Biology Specialist', true);

INSERT INTO space_mission.equipment_compatibility (primary_equipment_id, compatible_equipment_id, compatibility_level, notes) VALUES
(1, 2, 4, 'Oxygen tanks work well with solar panels'),
(1, 3, 2, 'Oxygen tanks have limited compatibility with rovers'),
(2, 1, 4, 'Solar panels work well with oxygen tanks'),
(2, 3, 5, 'Solar panels are optimal power source for rovers'),
(3, 2, 5, 'Rovers require solar panels for extended operation');

INSERT INTO space_mission.mission_status_history (
    mission_id, status, effective_from, effective_to, modified_by_astronaut_id, reason
) VALUES
(1, 'Planned', '2023-01-01 00:00:00', '2023-02-15 09:00:00', 1, 'Initial mission planning'),
(1, 'In Progress', '2023-02-15 09:00:01', '2023-05-20 14:30:00', 1, 'Mission launched'),
(1, 'Completed', '2023-05-20 14:30:01', NULL, 1, 'All objectives achieved'),
(2, 'Planned', '2023-01-15 00:00:00', '2023-03-10 08:15:00', 2, 'Initial mission planning'),
(2, 'In Progress', '2023-03-10 08:15:01', NULL, 2, 'Mission launched'),
(3, 'Planned', '2023-02-01 00:00:00', '2023-04-05 10:30:00', 3, 'Initial mission planning'),
(3, 'On Hold', '2023-04-05 10:30:01', '2023-04-20 16:45:00', 3, 'Technical issues'),
(3, 'In Progress', '2023-04-20 16:45:01', NULL, 3, 'Issues resolved, mission continuing');

INSERT INTO space_mission.equipment_status_history (
    equipment_id, mission_id, status, effective_from, effective_to, condition_rating, notes
) VALUES
(1, 1, 'Ready', '2023-01-15 00:00:00', '2023-02-15 09:00:00', 10, 'New equipment, pre-mission inspection passed'),
(1, 1, 'In Use', '2023-02-15 09:00:01', '2023-04-10 11:30:00', 9, 'Deployed for mission use'),
(1, 1, 'Maintenance', '2023-04-10 11:30:01', '2023-04-12 14:15:00', 7, 'Routine maintenance'),
(1, 1, 'In Use', '2023-04-12 14:15:01', NULL, 9, 'Back in service after maintenance'),
(2, 2, 'Ready', '2023-02-01 00:00:00', '2023-03-10 08:15:00', 10, 'New equipment'),
(2, 2, 'In Use', '2023-03-10 08:15:01', NULL, 8, 'Deployed for mission use'),
(3, 3, 'Ready', '2023-03-01 00:00:00', '2023-04-20 16:45:00', 10, 'New equipment'),
(3, 3, 'In Use', '2023-04-20 16:45:01', NULL, 9, 'Deployed after mission delay');

INSERT INTO space_mission.astronaut_role_history (
    astronaut_id, role, mission_id, effective_from, effective_to, supervisor_astronaut_id, performance_rating, notes
) VALUES
(3, 'Junior Engineer', 1, '2023-01-01 00:00:00', '2023-03-15 00:00:00', 1, 4, 'Initial assignment'),
(3, 'Lead Engineer', 1, '2023-03-15 00:00:01', NULL, 1, 5, 'Promotion due to exceptional performance'),
(4, 'Pilot', 2, '2023-01-15 00:00:00', NULL, 2, 4, 'Primary pilot for mission'),
(5, 'Science Officer', 3, '2023-02-01 00:00:00', '2023-04-10 00:00:00', 3, 3, 'Initial assignment'),
(5, 'Chief Science Officer', 3, '2023-04-10 00:00:01', NULL, 3, 4, 'Promotion after previous CSO reassignment');

INSERT INTO space_mission.astronaut_vitals (
    astronaut_id, mission_id, location, vital_data, recorded_at, is_anomalous
) VALUES
(3, 1, '(45.12, 22.45, 410.5)', 
   ROW(75, '120/80', 36.8, 98.5, '2023-03-10 09:15:00')::space_mission.vital_signs, 
   '2023-03-10 09:15:00', false),
(3, 1, '(45.12, 22.45, 410.5)', 
   ROW(82, '125/85', 36.9, 97.8, '2023-03-10 10:30:00')::space_mission.vital_signs, 
   '2023-03-10 10:30:00', false),
(4, 2, '(46.23, 23.56, 415.2)', 
   ROW(90, '130/85', 37.2, 96.5, '2023-03-15 14:20:00')::space_mission.vital_signs, 
   '2023-03-15 14:20:00', true),
(5, 3, '(44.89, 21.67, 408.9)', 
   ROW(72, '118/75', 36.7, 99.0, '2023-04-12 11:45:00')::space_mission.vital_signs, 
   '2023-04-12 11:45:00', false);
