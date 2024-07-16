INSERT INTO `m_db_1`.`container_status` (`id`) VALUES
(NULL), (NULL), (NULL), (NULL), (NULL), (NULL), (NULL), (NULL), (NULL), (NULL);

INSERT INTO `m_db_1`.`container` (`code`, `container_status_id`) VALUES
('code1', 4),
('code2', 1),
('code3', 3),
('code4', 1),
('code5', 5),
('code6', 2),
('code7', 9),
('code8', 8),
('code9', 6),
('code10',1);


INSERT INTO `m_db_2`.`container_status` (`id`) VALUES
(NULL), (NULL), (NULL), (NULL), (NULL), (NULL), (NULL), (NULL), (NULL), (NULL);

-- Insert 10 rows into container table
INSERT INTO `m_db_2`.`container` (`code`, `container_status_id`) VALUES
('code1', 1),
('code2', 4),
('code3', 5),
('code4', 8),
('code5', 3),
('code6', 1),
('code7', 9),
('code8', 7),
('code9', 2),
('code10', 1);
