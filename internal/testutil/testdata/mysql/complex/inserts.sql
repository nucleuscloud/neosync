SET FOREIGN_KEY_CHECKS=0;
-- =============================
-- 1) AGENCY (no dependencies)
-- =============================
INSERT INTO agency (name, country, founded_year)
VALUES
('NASA','USA',1958),
('ESA','Multinational',1975),
('Roscosmos','Russia',1992),
('CNSA','China',1993),
('JAXA','Japan',2003),
('ISRO','India',1969),
('SpaceX','USA',2002),
('Blue Origin','USA',2000),
('Rocket Lab','USA/NZ',2006),
('Arianespace','France',1980),
('CSA','Canada',1989),
('DLR','Germany',1997),
('UKSA','United Kingdom',2010),
('CNES','France',1961),
('ASI','Italy',1988),
('Inmarsat','United Kingdom',1979),
('KARI','South Korea',1989),
('UAE Space Agency','UAE',2014),
('Iran Space Agency','Iran',2004),
('Luxembourg Space Agency','Luxembourg',2018);

-- =========================================================
-- 2) CELESTIAL_BODY (self-reference via parent_body_id)
--    The first inserted row will get ID=1, second => ID=2, etc.
-- =========================================================
INSERT INTO celestial_body (name, body_type, mass, radius, parent_body_id)
VALUES
('Sun','Star',1.989e30,695700,NULL),         -- ID=1  (no parent)
('Mercury','Planet',3.301e23,2439.7,1),      -- ID=2  (parent=Sun)
('Venus','Planet',4.867e24,6051.8,1),        -- ID=3
('Earth','Planet',5.972e24,6371,1),          -- ID=4
('Mars','Planet',6.417e23,3389.5,1),         -- ID=5
('Jupiter','Planet',1.898e27,69911,1),       -- ID=6
('Saturn','Planet',5.683e26,58232,1),        -- ID=7
('Uranus','Planet',8.681e25,25362,1),        -- ID=8
('Neptune','Planet',1.024e26,24622,1),       -- ID=9
('Pluto','Dwarf Planet',1.303e22,1188.3,1),  -- ID=10
('Moon','Moon',7.3477e22,1737.4,4),          -- ID=11 (parent=Earth)
('Phobos','Moon',1.07e16,11.2667,5),         -- ID=12 (parent=Mars)
('Deimos','Moon',1.80e15,6.2,5),             -- ID=13
('Ganymede','Moon',1.495e23,2634.1,6),       -- ID=14
('Callisto','Moon',1.075e23,2410.3,6),       -- ID=15
('Titan','Moon',1.345e23,2575.5,7),          -- ID=16
('Enceladus','Moon',1.08e20,252.1,7),        -- ID=17
('Europa','Moon',4.80e22,1560.8,6),          -- ID=18
('Halley','Comet',2.2e14,11.0,1),            -- ID=19
('Ceres','Dwarf Planet',9.39e20,473,1);      -- ID=20

-- ===========================================
-- 3) LAUNCH_SITE (no references to other new tables)
--    but has a spatial POINT column
-- ===========================================
INSERT INTO launch_site (name, location, location_coord, country)
VALUES
('Cape Canaveral LC-39A','Florida, USA', ST_GeomFromText('POINT(-80.6042 28.6084)'), 'USA'),
('Cape Canaveral LC-40','Florida, USA', ST_GeomFromText('POINT(-80.6019 28.5614)'), 'USA'),
('Vandenberg SLC-4E','California, USA', ST_GeomFromText('POINT(-120.6106 34.6321)'), 'USA'),
('Vandenberg SLC-4W','California, USA', ST_GeomFromText('POINT(-120.6136 34.6331)'), 'USA'),
('Baikonur Cosmodrome','Baikonur, Kazakhstan', ST_GeomFromText('POINT(63.3422 45.9206)'), 'Kazakhstan'),
('Plesetsk Cosmodrome','Plesetsk, Russia', ST_GeomFromText('POINT(40.6469 62.9270)'), 'Russia'),
('Guiana Space Centre ELA-3','Kourou, French Guiana', ST_GeomFromText('POINT(-52.7750 5.2360)'), 'France'),
('Guiana Space Centre ELS','Kourou, French Guiana', ST_GeomFromText('POINT(-52.7685 5.2220)'), 'France'),
('Jiuquan SLS-2','Jiuquan, China', ST_GeomFromText('POINT(100.2916 40.9583)'), 'China'),
('Satish Dhawan SLP','Sriharikota, India', ST_GeomFromText('POINT(80.2300 13.7330)'), 'India'),
('Tanegashima Yoshinobu','Tanegashima, Japan', ST_GeomFromText('POINT(130.966 30.405)'), 'Japan'),
('Wenchang LC-101','Wenchang, China', ST_GeomFromText('POINT(110.951 19.614)'), 'China'),
('Kennedy LC-39B','Florida, USA', ST_GeomFromText('POINT(-80.6215 28.6270)'), 'USA'),
('Blue Origin Launch Site One','West Texas, USA', ST_GeomFromText('POINT(-104.757 31.400)'), 'USA'),
('Rocket Lab Launch Complex 1','Mahia Peninsula, NZ', ST_GeomFromText('POINT(177.864 -39.262)'), 'New Zealand'),
('Kodiak Launch Complex','Alaska, USA', ST_GeomFromText('POINT(-152.337 57.436)'), 'USA'),
('JAXA Uchinoura','Kagoshima, Japan', ST_GeomFromText('POINT(131.079 31.251)'), 'Japan'),
('Palmachim','Palmachim, Israel', ST_GeomFromText('POINT(34.700 31.900)'), 'Israel'),
('Xichang','Xichang, China', ST_GeomFromText('POINT(102.025 28.246)'), 'China'),
('Alcantara','Alcantara, Brazil', ST_GeomFromText('POINT(-44.396 2.307)'), 'Brazil');

-- ==========================================================
-- 4) OBSERVATORY (references agency.id, launch_site.id)
--    Some will have launch_site_id = NULL
-- ==========================================================
INSERT INTO observatory (name, agency_id, launch_site_id, location_coord, status)
VALUES
('Kitt Peak National Observatory', 1, NULL, ST_GeomFromText('POINT(-111.6003 31.9583)'), 'Active'),
('Paranal Observatory', 2, NULL, ST_GeomFromText('POINT(-70.4030 -24.6252)'), 'Active'),
('La Silla Observatory', 2, NULL, ST_GeomFromText('POINT(-70.7345 -29.2570)'), 'Active'),
('Very Large Array', 1, NULL, ST_GeomFromText('POINT(-107.6184 34.0784)'), 'Under Maintenance'),
('Mount Wilson Observatory', 1, NULL, ST_GeomFromText('POINT(-118.0662 34.2242)'), 'Active'),
('Guiana Space Centre Observatory', 10, 7, ST_GeomFromText('POINT(-52.7670 5.2320)'), 'Active'),
('Baikonur Observatory', 3, 5, ST_GeomFromText('POINT(63.3425 45.9210)'), 'Active'),
('Jiuquan Observatory', 4, 9, ST_GeomFromText('POINT(100.2890 40.9600)'), 'Active'),
('Satish Dhawan Observatory', 6, 10, ST_GeomFromText('POINT(80.2310 13.7340)'), 'Active'),
('Tanegashima Observatory', 5, 11, ST_GeomFromText('POINT(130.970 30.403)'), 'Active'),
('Vandenberg Observatory', 1, 3, ST_GeomFromText('POINT(-120.6110 34.6300)'), 'Active'),
('Wenchang Observatory', 4, 12, ST_GeomFromText('POINT(110.950 19.615)'), 'Active'),
('Kennedy Observatory', 1, 13, ST_GeomFromText('POINT(-80.6200 28.6260)'), 'Under Maintenance'),
('Blue Origin Observatory', 8, 14, NULL, 'Active'),
('Rocket Lab Mahia Observatory', 9, 15, NULL, 'Active'),
('Kodiak Observatory', 1, 16, NULL, 'Active'),
('JAXA Uchinoura Observatory', 5, 17, NULL, 'Active'),
('Palmachim Observatory', 1, 18, NULL, 'Active'),
('Xichang Observatory', 4, 19, NULL, 'Active'),
('Alcantara Observatory', 1, 20, NULL, 'Active');

-- ===========================================
-- 5) TELESCOPE (references observatory.id)
--    Must respect chk_telescope_mirror constraint
-- ===========================================
INSERT INTO telescope (observatory_id, name, telescope_type, mirror_diameter_m, status)
VALUES
(1, 'KPNO 2.1m Reflector', 'Optical', 2.1, 'Operational'),
(2, 'VLT UT1', 'Optical', 8.2, 'Operational'),
(3, 'La Silla 3.6m', 'Optical', 3.6, 'Operational'),
(4, 'VLA Antenna 1', 'Radio', NULL, 'Operational'),
(4, 'VLA Antenna 2', 'Radio', NULL, 'Operational'),
(5, 'Mt. Wilson Hooker', 'Optical', 2.5, 'Damaged'),
(6, 'Guiana Tracking Scope', 'Optical', 1.0, 'Operational'),
(7, 'Baikonur Radio Scope', 'Radio', NULL, 'Operational'),
(8, 'Jiuquan Infrared Telescope', 'Infrared', 1.2, 'Operational'),
(9, 'Satish Radio Array', 'Radio', NULL, 'Operational'),
(10, 'Tanegashima X-Ray Telescope', 'X-Ray', 1.5, 'Operational'),
(11, 'Vandenberg Surveillance Scope', 'Optical', 0.8, 'Operational'),
(12, 'Wenchang UV Scope', 'UV', 2.0, 'Operational'),
(13, 'Kennedy Multi-Purpose Scope', 'Other', NULL, 'Operational'),
(14, 'Blue Origin Radio Net', 'Radio', NULL, 'Operational'),
(15, 'Rocket Lab Tracking Scope', 'Optical', 0.5, 'Damaged'),
(16, 'Kodiak Optical', 'Optical', 0.9, 'Operational'),
(17, 'Uchinoura Space Monitor', 'Optical', 1.6, 'Operational'),
(18, 'Palmachim Optical', 'Optical', 2.0, 'Operational'),
(19, 'Xichang Radio Array', 'Radio', NULL, 'Operational'),
(20, 'Alcantara Optical', 'Optical', 1.5, 'Operational');

-- ====================================
-- 6) SPACECRAFT (references agency.id)
-- ====================================
INSERT INTO spacecraft (name, type, capacity, status, agency_id, last_mission_id)
VALUES
('Crew Dragon','Crewed',4,'Operational',7,NULL),
('Starliner','Crewed',4,'Operational',7,NULL),
('Soyuz MS','Crewed',3,'Operational',3,NULL),
('Shenzhou','Crewed',3,'Operational',4,NULL),
('Orion','Crewed',4,'Operational',1,NULL),
('Hubble Space Telescope','Orbiter',0,'Operational',1,NULL),
('Voyager 1','Probe',0,'Retired',1,NULL),
('Voyager 2','Probe',0,'Retired',1,NULL),
('Perseverance Rover','Rover',0,'Operational',1,NULL),
('Apollo CSM','Crewed',3,'Retired',1,NULL),
('Tiangong Station','Station',6,'Operational',4,NULL),
('ISS Module','Station',6,'Operational',1,NULL),
('BFS Starship','Crewed',100,'In Mission',7,NULL),
('Blue Moon Lander','Lander',0,'Operational',8,NULL),
('New Shepard','Crewed',6,'Operational',8,NULL),
('Electron Photon','Orbiter',0,'Operational',9,NULL),
('Falcon Heavy Upper Stage','Orbiter',0,'Operational',7,NULL),
('PSLV Orbiter','Orbiter',0,'Operational',6,NULL),
('Ariane Transfer Vehicle','Orbiter',0,'Operational',10,NULL),
('X-37B','Crewed',2,'Operational',1,NULL);

-- =======================================================
-- 7) ASTRONAUT (references agency.id, first_mission_id later)
--    We'll set first_mission_id = NULL initially
-- =======================================================
INSERT INTO astronaut (first_name, last_name, birth_date, nationality, status, agency_id, first_mission_id)
VALUES
('Neil','Armstrong','1930-08-05','American','Deceased',1,NULL),
('Buzz','Aldrin','1930-01-20','American','Retired',1,NULL),
('Yuri','Gagarin','1934-03-09','Russian','Deceased',3,NULL),
('Valentina','Tereshkova','1937-03-06','Russian','Retired',3,NULL),
('Sally','Ride','1951-05-26','American','Deceased',1,NULL),
('Chris','Hadfield','1959-08-29','Canadian','Retired',11,NULL),
('Samantha','Cristoforetti','1977-04-26','Italian','Active',15,NULL),
('Yang','Liwei','1965-06-21','Chinese','Active',4,NULL),
('Sunita','Williams','1965-09-19','American','Active',1,NULL),
('Tim','Peake','1972-04-07','British','Active',13,NULL),
('Alexander','Ger-st','1976-05-03','German','Active',12,NULL),
('Peggy','Whitson','1960-02-09','American','Retired',1,NULL),
('Kate','Rubins','1978-10-14','American','Active',1,NULL),
('Akihiko','Hoshide','1968-12-28','Japanese','Active',5,NULL),
('Rakesh','Sharma','1949-01-13','Indian','Retired',6,NULL),
('Mae','Jemison','1956-10-17','American','Retired',1,NULL),
('Luca','Parmitano','1976-09-27','Italian','Active',15,NULL),
('Mark','Kelly','1964-02-21','American','Active',1,NULL),
('Oleg','Kononenko','1964-06-21','Russian','Active',3,NULL),
('Jessica','Meir','1977-07-01','American','Active',1,NULL);

-- =============================================================
-- 8) MISSION
--    References: spacecraft_id, destination_id, launch_site_id,
--                primary_agency_id, commander_id (only if manned)
--    We'll create 20. Some manned (commander_id NOT NULL), some unmanned.
-- =============================================================
INSERT INTO mission
(name, mission_code, mission_type, status, launch_date, return_date,
 spacecraft_id, destination_id, launch_site_id, primary_agency_id, commander_id)
VALUES
('Apollo 11','M-001','Manned','Completed','1969-07-16','1969-07-24',10,4,1,1,1), 
('Soyuz TMA-1','M-002','Manned','Completed','2002-10-30','2003-05-04',3,4,5,3,4),
('Shenzhou 5','M-003','Manned','Completed','2003-10-15','2003-10-16',4,4,9,4,8),
('ISS Expedition 50','M-004','Manned','Completed','2016-10-17','2017-04-10',12,4,3,1,11),
('Crew-1','M-005','Manned','Completed','2020-11-16','2021-05-02',1,4,1,1,9),
('Crew-2','M-006','Manned','Completed','2021-04-23','2021-11-08',1,4,13,1,18),
('Gaganyaan Demo','M-007','Manned','Planned','2025-01-01',NULL,18,4,10,6,15),
('JAXA Hoshide Flight','M-008','Manned','Active','2023-06-10',NULL,5,4,11,5,14),
('ESA Cristoforetti Flight','M-009','Manned','Active','2022-04-01',NULL,2,4,3,2,7),
('Roscosmos Kononenko Flight','M-010','Manned','Planned','2024-10-10',NULL,3,4,6,3,19),
('Voyager 1 Launch','M-011','Unmanned','Completed','1977-09-05','1977-09-05',7,1,1,1,NULL),
('Voyager 2 Launch','M-012','Unmanned','Completed','1977-08-20','1977-08-20',8,1,1,1,NULL),
('Perseverance to Mars','M-013','Unmanned','Active','2020-07-30',NULL,9,5,1,1,NULL),
('Chang\'e 4','M-014','Unmanned','Completed','2018-12-08','2019-01-03',4,11,9,4,NULL),
('Tianwen-1','M-015','Unmanned','Active','2020-07-23',NULL,4,5,9,4,NULL),
('Mangalyaan','M-016','Unmanned','Completed','2013-11-05','2014-09-24',18,5,10,6,NULL),
('New Shepard Test','M-017','Unmanned','Completed','2019-01-23','2019-01-23',15,1,14,8,NULL),
('Electron Demo','M-018','Unmanned','Completed','2017-05-25','2017-05-25',16,1,15,9,NULL),
('Ariane Transfer Test','M-019','Unmanned','Completed','2019-10-01','2019-10-01',19,1,7,10,NULL),
('X-37B OTV-6','M-020','Unmanned','Active','2020-05-17',NULL,20,1,3,1,NULL);

-- =================================================================
-- 9) MISSION_CREW (references mission.id and astronaut.id)
--    We only assign crew to MANNED missions (IDs 1..10 of mission).
--    We'll insert exactly 20 rows: 2 for each manned mission.
-- =================================================================
INSERT INTO mission_crew (mission_id, astronaut_id, role)
VALUES
-- Apollo 11
(1, 1, 'Commander'),
(1, 2, 'Pilot'),
-- Soyuz TMA-1
(2, 4, 'Commander'),
(2, 19, 'Engineer'),
-- Shenzhou 5
(3, 8, 'Commander'),
(3, 7, 'Scientist'),
-- ISS Expedition 50
(4, 11, 'Commander'),
(4, 20, 'Engineer'),
-- Crew-1
(5, 9, 'Commander'),
(5, 13, 'Engineer'),
-- Crew-2
(6, 18, 'Commander'),
(6, 10, 'Scientist'),
-- Gaganyaan Demo
(7, 15, 'Commander'),
(7, 5, 'Scientist'),
-- JAXA Hoshide Flight
(8, 14, 'Commander'),
(8, 6, 'Engineer'),
-- ESA Cristoforetti Flight
(9, 7, 'Commander'),
(9, 16, 'Engineer'),
-- Roscosmos Kononenko Flight
(10, 19, 'Commander'),
(10, 3, 'Specialist');

-- =====================================
-- 10) MISSION_LOG (references mission.id)
--     20 log entries
-- =====================================
INSERT INTO mission_log (mission_id, log_time, event)
VALUES
(1, '2006-07-16 13:32:00', 'Launch from Kennedy LC-39A'),
(1, '2006-07-19 18:00:00', 'Lunar orbit insertion'),
(1, '2016-07-21 02:56:00', 'First step on Moon'),
(2, '2002-10-30 06:00:00', 'Launch successful'),
(2, '2003-05-04 02:00:00', 'Landing in Kazakhstan'),
(3, '2003-10-15 09:00:00', 'Launch from Jiuquan'),
(4, '2016-10-17 10:00:00', 'Expedition 50 started'),
(4, '2017-04-10 21:00:00', 'Expedition 50 ended'),
(5, '2020-11-16 19:27:00', 'Crew-1 Liftoff'),
(5, '2021-05-02 06:35:00', 'Crew-1 Splashdown'),
(6, '2021-04-23 09:49:00', 'Crew-2 Launch'),
(7, '2025-01-01 05:00:00', 'Gaganyaan Demo scheduled'),
(8, '2023-06-10 07:15:00', 'JAXA Hoshide Flight launched'),
(9, '2022-04-01 08:00:00', 'ESA Cristoforetti Flight begun'),
(10, '2024-10-10 06:00:00', 'Roscosmos Kononenko flight scheduled'),
(11, '2016-09-05 14:56:00', 'Voyager 1 Launch'),
(12, '2016-08-20 12:29:00', 'Voyager 2 Launch'),
(13, '2020-07-30 11:50:00', 'Perseverance launched to Mars'),
(14, '2018-12-08 18:23:00', 'Chang\'e 4 launched'),
(15, '2020-07-23 12:41:00', 'Tianwen-1 launched');

-- =============================================================
-- 11) RESEARCH_PROJECT (references lead_astronaut_id)
--     20 distinct projects
-- =============================================================
INSERT INTO research_project (title, description, start_date, end_date, lead_astronaut_id)
VALUES
('Lunar Soil Analysis','Study of lunar regolith samples','2021-01-01','2021-12-31',2),
('Mars Rover Experiments','Analyze data from Perseverance','2021-02-15','2022-06-30',9),
('ISS Plant Growth','Growing plants in microgravity','2019-05-01','2020-08-15',11),
('Solar Flare Observations','Tracking solar flare activity','2020-01-01',NULL,7),
('Space Medicine Trial','Effects of microgravity on bone density','2022-01-01',NULL,10),
('Lunar Water Mapping','Remote sensing for lunar water','2023-03-01',NULL,7),
('Asteroid Mining Feasibility','Mining operations in near-earth asteroids','2024-01-10',NULL,8),
('Exoplanet Atmosphere Study','Spectroscopic analysis of exoplanets','2020-06-01','2022-06-01',14),
('Radiation Shielding Tests','Testing new materials for cosmic rays','2021-09-01','2023-03-01',16),
('Deep Space Habitats','Designing habitats for long-duration missions','2021-11-01',NULL,18),
('Cislunar Navigation Systems','Testing navigation near the Moon','2022-04-10','2022-12-31',4),
('Space Debris Tracking','Tracking and cataloging orbital debris','2021-10-01',NULL,20),
('MOXIE Enhancement','Improving oxygen production on Mars','2023-07-01',NULL,15),
('Europa Subsurface Probe','Concept for exploring Europa','2025-01-01',NULL,19),
('Multinational Lunar Station','Designing an international lunar outpost','2023-05-05',NULL,7),
('Microbe Studies in Orbit','Microorganisms behavior in LEO','2024-02-02','2024-12-31',9),
('High-Definition Earth Imaging','Advanced imaging from ISS','2022-08-15','2023-09-10',5),
('Plasma Physics in Microgravity','Plasma experiments in orbit','2021-01-10','2022-01-10',8),
('Asteroid Redirect Mission','Test mission for redirecting asteroids','2023-09-01',NULL,3),
('Lunar Gateway Tech Demo','Tech demos for Gateway station','2022-01-05',NULL,1);

-- ==================================================================
-- 12) PROJECT_MISSION (references research_project.id, mission.id)
--     20 combos (no duplicates)
-- ==================================================================
INSERT INTO project_mission (project_id, mission_id)
VALUES
(1, 1),
(2, 13),
(3, 4),
(4, 11),
(5, 5),
(6, 1),
(7, 16),
(8, 18),
(9, 4),
(10, 6),
(11, 14),
(12, 20),
(13, 13),
(14, 9),
(15, 1),
(16, 5),
(17, 4),
(18, 8),
(19, 12),
(20, 2);

-- ==============================================================
-- 13) INSTRUMENT (references telescope_id, spacecraft_id)
--     20 instruments, some on telescopes, some on spacecraft
-- ==============================================================
INSERT INTO instrument (name, instrument_type, telescope_id, spacecraft_id, status)
VALUES
('Wide Field Camera','Camera',1,NULL,'In Use'),
('Spectro Analyzer','Spectrometer',2,NULL,'Available'),
('Thermal Sensor','Sensor',3,NULL,'Damaged'),
('Deep Space Cam','Camera',4,NULL,'Available'),
('UV Spectrometer','Spectrometer',5,NULL,'Available'),
('InfraRed Mapper','Camera',9,NULL,'Available'),
('X-Ray Detector','Sensor',10,NULL,'Available'),
('Orbit Docking Module','Module',NULL,10,'In Use'),
('High-Gain Antenna','Sensor',NULL,7,'Available'),
('Mars Drill','Other',NULL,9,'In Use'),
('CO2 Scrubber','Module',NULL,5,'Available'),
('Lunar Lander Cam','Camera',NULL,14,'Available'),
('Starship Life Support','Module',NULL,13,'In Use'),
('ISS Research Module','Module',NULL,12,'In Use'),
('Soyuz Docking Adapter','Other',NULL,3,'Available'),
('Crew Dragon Display','Other',NULL,1,'Available'),
('Probe Thermal Shield','Module',NULL,8,'Retired'),
('Rover Soil Sensor','Sensor',NULL,9,'In Use'),
('Radio Frequency Sensor','Sensor',19,NULL,'Available'),
('Navigation Beacon','Other',NULL,19,'Available');

-- ==================================================================
-- 14) OBSERVATION_SESSION (refs telescope_id, instrument_id, 
--                          target_body_id, mission_id)
--     20 sessions. Avoid referencing 'Retired' telescopes or 
--     obviously invalid combos.
-- ==================================================================
INSERT INTO observation_session
(telescope_id, instrument_id, target_body_id, mission_id, start_time, end_time, seeing_conditions, notes)
VALUES
(1, 1, 1, NULL, '2021-01-01 03:00:00','2021-01-01 06:00:00','Excellent','Solar observation at dawn'),
(2, 2, 3, NULL, '2022-02-10 20:00:00','2022-02-11 02:00:00','Good','Venus spectroscopy'),
(3, 3, 2, NULL, '2022-03-15 22:00:00','2022-03-16 01:00:00','Fair','Mercury thermal scan'),
(4, 4, 1, NULL, '2023-01-05 18:00:00','2023-01-05 20:00:00','Good','Radio check of sun activity'),
(9, 6, 5, 13, '2021-08-01 10:00:00','2021-08-01 12:00:00','Excellent','Mars mapping from orbit'),
(10, 7, 11, 14, '2019-01-03 00:00:00','2019-01-03 02:00:00','Poor','Chang\'e 4 lunar landing observation'),
(11, 1, 4, 2, '2002-10-30 07:00:00','2002-10-30 08:30:00','Good','Earth coverage during Soyuz TMA-1'),
(12, 2, 4, 6, '2021-04-23 10:00:00','2021-04-23 11:00:00','Good','Crew-2 Earth observation'),
(14, 5, 1, 17, '2019-01-23 10:15:00','2019-01-23 10:45:00','Excellent','New Shepard test flight tracking'),
(16, 1, 1, NULL, '2022-07-10 22:00:00','2022-07-10 23:00:00','Fair','Night sky solar calibration?'),
(17, 2, 4, 8, '2023-06-10 08:30:00','2023-06-10 09:30:00','Good','Hoshide flight Earth imaging'),
(18, 4, 1, 9, '2022-04-01 10:00:00','2022-04-01 10:30:00','Excellent','Cristoforetti flight solar check'),
(19, 19, 1, NULL, '2023-01-15 06:00:00','2023-01-15 08:00:00','Good','Radio array scanning sun'),
(20, NULL, 4, NULL, '2023-04-20 12:00:00','2023-04-20 13:00:00','Fair','Earth imaging test'),
(2, NULL, 9, NULL, '2023-05-10 22:00:00','2023-05-10 23:00:00','Poor','Neptune test pointing'),
(5, NULL, 4, NULL, '2022-06-11 09:00:00','2022-06-11 09:15:00','Good','Quick Earth check'),
(9, 10, 5, 16, '2013-11-05 05:10:00','2013-11-05 06:00:00','Excellent','Mangalyaan launch tracking'),
(3, 2, 19, NULL, '2023-06-01 19:00:00','2023-06-01 19:30:00','Good','Halley Comet radio check'),
(10, 7, 11, NULL, '2020-12-01 02:00:00','2020-12-01 03:00:00','Fair','Moon x-ray imaging test'),
(11, 1, 4, 5, '2020-11-16 20:00:00','2020-11-16 21:00:00','Good','Crew-1 Earth observation');

-- ==============================================================
-- 15) DATA_SET (references mission_id, observation_id)
--     20 data sets
-- ==============================================================
INSERT INTO data_set
(name, mission_id, observation_id, data_description, data_blob, collected_on)
VALUES
('Apollo 11 EVA Photos', 1, NULL, 'Images captured during lunar EVA', NULL, '1969-07-21'),
('Soyuz TMA-1 Reentry Data', 2, 7, 'Telemetry from reentry phase', NULL, '2003-05-04'),
('Shenzhou 5 Launch Telemetry', 3, NULL, 'Launch-phase telemetry logs', NULL, '2003-10-16'),
('VLA Sun Scan 2023', NULL, 4, 'Radio observations of sun activity', NULL, '2023-01-05'),
('Perseverance Mars Images', 13, 5, 'High-res color images of Mars surface', NULL, '2021-08-01'),
('Chang\'e 4 Landing Data', 14, 6, 'Lunar far side landing data', NULL, '2019-01-03'),
('Crew-1 Earth Photos', 5, 20, 'Photos of Earth during Crew-1 mission', NULL, '2020-11-16'),
('Crew-2 Solar Observations', 6, 12, 'Sun observations from orbit', NULL, '2021-04-23'),
('New Shepard Flight Test Data', 17, 9, 'Flight telemetry of suborbital test', NULL, '2019-01-23'),
('Neptune Spectral Data', NULL, 15, 'Spectral lines measurement at Neptune', NULL, '2023-05-10'),
('Mars Launch Tracking', 16, 17, 'Launch tracking data for Mangalyaan', NULL, '2013-11-05'),
('Halley Comet Radio Observations', NULL, 18, 'Comet radio signature logs', NULL, '2023-06-01'),
('Moon X-Ray Test', NULL, 19, 'Preliminary x-ray imaging of moon', NULL, '2020-12-01'),
('Gaganyaan Preliminary Data', 7, NULL, 'Planned data sets for Gaganyaan Demo', NULL, '2025-01-01'),
('Apollo 11 Sample Analysis', 1, NULL, 'Follow-up on lunar rock composition', NULL, '1969-08-01'),
('Crew-2 Docking Logs', 6, NULL, 'Data logs from ISS docking operations', NULL, '2021-04-24'),
('Voyager 1 Launch Video', 11, NULL, 'Historic launch footage', NULL, '1977-09-05'),
('Voyager 2 Launch Telemetry', 12, NULL, 'Launch data logs', NULL, '1977-08-20'),
('X-37B Flight Data', 20, NULL, 'OTV-6 in-orbit experiment logs', NULL, '2020-05-17'),
('ISS Exp 50 Plant Growth Logs', 4, 8, 'Plant growth experiment data on ISS', NULL, '2016-10-18');

-- ==================================================================
-- 16) RESEARCH_PAPER (references project_id, observatory_id)
--     20 papers
-- ==================================================================
INSERT INTO research_paper
(title, abstract, published_date, doi, project_id, observatory_id)
VALUES
('Lunar Soil Composition','Study of Apollo 11 soil samples','1970-01-10','10.1234/lunar.1970',1,1),
('Mars Rover Camera Analysis','Review of Perseverance imaging tech','2021-12-01','10.1234/mars.2021',2,2),
('ISS Plant Growth Results','Plant experiments in microgravity','2020-01-15','10.1234/iss.2020',3,4),
('Solar Flare Impact','Observations of solar flare activity','2020-08-01','10.1234/sol.2020',4,1),
('Space Medicine Advances','Bone density and microgravity','2022-03-01','10.1234/med.2022',5,4),
('Lunar Water Prospecting','Findings on remote sensing of lunar water','2024-01-01','10.1234/lunwat.2024',6,2),
('Asteroid Mining Potential','Tech feasibility of asteroid resources','2025-04-10','10.1234/astmine.2025',7,3),
('Exoplanet Atmospheres','Spectroscopic approaches','2021-07-01','10.1234/exo.2021',8,3),
('Radiation Shield Test','Lab results for cosmic ray shielding','2023-04-15','10.1234/rad.2023',9,4),
('Deep Space Habitat Concepts','Architectures for long missions','2022-11-20','10.1234/dsh.2022',10,5),
('Cislunar Nav & Guidance','Methods for lunar orbit nav','2023-02-10','10.1234/nav.2023',11,6),
('Space Debris Mitigation','Tech to reduce orbital debris','2021-11-11','10.1234/debris.2021',12,7),
('Mars Oxygen ISRU','Enhancements to MOXIE','2024-06-22','10.1234/moxie.2024',13,8),
('Europa Probe Design','Conceptual design for subsurface probe','2025-02-01','10.1234/europa.2025',14,9),
('Lunar Station Architecture','Plans for multinational station','2023-07-07','10.1234/lunast.2023',15,10),
('Microbes in Orbit','Results from microorganisms in LEO','2025-06-15','10.1234/microbe.2025',16,11),
('HD Earth Imaging','Tech used for ISS imaging','2023-02-28','10.1234/hdearth.2023',17,12),
('Plasma Experiments','Plasma behavior in microgravity','2022-05-05','10.1234/plasma.2022',18,13),
('Asteroid Redirect Feasibility','Examining methods to deflect asteroids','2024-01-20','10.1234/arm.2024',19,14),
('Lunar Gateway Tech Preview','Early results from gateway prototypes','2023-03-01','10.1234/gateway.2023',20,15);

-- =====================================================================
-- 17) PAPER_CITATION (references research_paper.id)
--     20 citations, no self-citations, no duplicates
-- =====================================================================
INSERT INTO paper_citation (citing_paper_id, cited_paper_id, citation_date)
VALUES
(2,1,'2021-12-10'),
(3,1,'2020-01-20'),
(4,1,'2020-08-10'),
(5,1,'2022-03-05'),
(5,2,'2022-03-06'),
(6,4,'2024-01-02'),
(7,4,'2025-04-15'),
(8,2,'2021-07-10'),
(9,2,'2023-04-20'),
(10,3,'2022-11-30'),
(11,1,'2023-02-15'),
(12,9,'2021-11-12'),
(13,2,'2024-07-01'),
(14,4,'2025-02-10'),
(15,6,'2023-07-08'),
(16,3,'2025-06-20'),
(17,3,'2023-03-01'),
(18,9,'2022-05-10'),
(19,7,'2024-01-25'),
(20,1,'2023-03-05');

-- ============================================
-- 18) GRANT (references agency.id)
--     20 grants
-- ============================================
INSERT INTO `grant` (grant_number, agency_id, funding_amount, start_date, end_date, status)
VALUES
('GRANT-001', 1, 1000000.00, '2020-01-01','2020-12-31','Closed'),
('GRANT-002', 2, 2000000.00, '2021-01-01','2022-01-01','Closed'),
('GRANT-003', 3, 500000.00, '2022-01-01','2023-01-01','Closed'),
('GRANT-004', 4, 2500000.00, '2023-01-01',NULL,'Active'),
('GRANT-005', 5, 750000.00, '2021-06-01','2022-06-01','Closed'),
('GRANT-006', 6, 300000.00, '2022-05-01','2023-04-30','Closed'),
('GRANT-007', 7, 15000000.00, '2023-01-01','2025-12-31','Active'),
('GRANT-008', 8, 5000000.00, '2022-11-01','2024-11-01','Active'),
('GRANT-009', 9, 10000000.00, '2020-02-01',NULL,'Awarded'),
('GRANT-010', 10, 20000000.00, '2019-01-01',NULL,'Canceled'),
('GRANT-011', 11, 400000.00, '2021-04-01','2023-04-01','Closed'),
('GRANT-012', 12, 1000000.00, '2022-02-10',NULL,'Active'),
('GRANT-013', 13, 500000.00, '2023-03-10',NULL,'Proposed'),
('GRANT-014', 14, 7500000.00, '2022-07-01','2023-07-01','Closed'),
('GRANT-015', 15, 250000.00, '2023-01-10','2023-12-31','Active'),
('GRANT-016', 16, 125000.00, '2021-01-05','2022-01-05','Closed'),
('GRANT-017', 17, 100000000.00, '2024-01-01','2029-12-31','Proposed'),
('GRANT-018', 18, 2000000.00, '2023-06-01',NULL,'Active'),
('GRANT-019', 19, 300000.00, '2022-01-20','2022-12-20','Closed'),
('GRANT-020', 20, 999999.99, '2024-02-02',NULL,'Proposed');

-- ======================================================================
-- 19) GRANT_RESEARCH_PROJECT (references grant.id, research_project.id)
--     20 combos
-- ======================================================================
INSERT INTO grant_research_project (grant_id, research_project_id, allocated_amount)
VALUES
(1,1,100000.00),
(2,2,1500000.00),
(3,3,300000.00),
(4,4,2000000.00),
(5,5,500000.00),
(6,6,250000.00),
(7,7,10000000.00),
(8,8,4500000.00),
(9,9,9000000.00),
(10,10,18000000.00),
(11,11,200000.00),
(12,12,900000.00),
(13,13,400000.00),
(14,14,7000000.00),
(15,15,100000.00),
(16,16,100000.00),
(17,17,50000000.00),
(18,18,1500000.00),
(19,19,200000.00),
(20,20,500000.00);

-- ======================================================================
-- 20) INSTRUMENT_USAGE (references instrument.id, telescope.id, spacecraft.id)
--     20 usage entries
-- ======================================================================
INSERT INTO instrument_usage
(instrument_id, telescope_id, spacecraft_id, start_date, end_date, usage_notes)
VALUES
(1, 1, NULL, '2021-01-01','2021-02-01','Camera used on KPNO telescope'),
(2, 2, NULL, '2021-06-01','2021-06-10','Spectrometer trial on VLT UT1'),
(3, 3, NULL, '2022-03-15','2022-03-20','Thermal sensor test at La Silla'),
(4, 4, NULL, '2023-01-05',NULL,'Deep Space Cam on VLA Antenna 1'),
(5, 5, NULL, '2023-06-10',NULL,'UV spectroscopy test at Mt. Wilson'),
(6, 9, NULL, '2021-08-01','2021-08-02','IR Mapper for Mars observation'),
(7, 10, NULL, '2019-01-03','2019-01-04','X-Ray detection from Chang\'e 4 vantage'),
(8, NULL, 10, '2021-05-01','2021-05-15','Docking module installed on Apollo CSM'),
(9, NULL, 7, '1977-09-05','1977-09-10','High-Gain Antenna for Voyager 1'),
(10, NULL, 9, '2020-07-30','2020-08-15','Mars Drill for Perseverance'),
(11, NULL, 5, '2023-06-10','2023-06-20','CO2 Scrubber test on Orion'),
(12, NULL, 14, '2020-12-01','2020-12-10','Lunar Lander Cam test on Blue Moon'),
(13, NULL, 13, '2023-02-01','2023-03-01','Life support system usage on Starship'),
(14, NULL, 12, '2016-10-17','2017-04-10','ISS research module usage in Exp 50'),
(15, NULL, 3, '2002-10-30','2002-10-31','Soyuz docking adapter test'),
(16, NULL, 1, '2020-11-16','2020-11-20','Crew Dragon display test'),
(17, NULL, 8, '1977-08-20','1977-08-28','Probe thermal shield on Voyager 2'),
(18, NULL, 9, '2021-08-01','2021-09-01','Soil sensor on Perseverance'),
(19, 19, NULL, '2023-01-15','2023-01-16','Radio frequency sensor on Xichang array'),
(20, NULL, 19, '2019-10-01','2019-10-02','Navigation beacon on Ariane Transfer Vehicle');


UPDATE astronaut SET first_mission_id = 1 WHERE id IN (1, 2);   -- Armstrong, Aldrin
UPDATE astronaut SET first_mission_id = 2 WHERE id IN (4, 19);  -- Tereshkova, Kononenko
UPDATE astronaut SET first_mission_id = 3 WHERE id IN (7, 8);   -- Cristoforetti, Yang
UPDATE astronaut SET first_mission_id = 4 WHERE id IN (11, 20); -- Gerst, Meir
UPDATE astronaut SET first_mission_id = 5 WHERE id IN (9, 13);  -- Williams, Rubins
UPDATE astronaut SET first_mission_id = 6 WHERE id IN (10, 18); -- Peake, Mark Kelly
UPDATE astronaut SET first_mission_id = 7 WHERE id IN (5, 15);  -- Ride, Rakesh
UPDATE astronaut SET first_mission_id = 8 WHERE id IN (6, 14);  -- Hadfield, A. Hoshide
UPDATE astronaut SET first_mission_id = 9 WHERE id = 16;         -- Mae Jemison
UPDATE astronaut SET first_mission_id = 10 WHERE id = 3;         -- Yuri Gagarin
SET FOREIGN_KEY_CHECKS=1;

