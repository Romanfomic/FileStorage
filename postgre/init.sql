CREATE TABLE Permissions (
    permission_id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    description TEXT
);

CREATE TABLE Roles (
    role_id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    description TEXT
);

CREATE TABLE Role_Permissions (
    role_id INTEGER REFERENCES Roles(role_id) ON DELETE CASCADE,
    permission_id INTEGER REFERENCES Permissions(permission_id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE Groups (
    group_id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    description TEXT
);

CREATE TABLE Access (
    access_id SERIAL PRIMARY KEY,
    name VARCHAR(100)
);

CREATE TABLE Users (
    user_id SERIAL PRIMARY KEY,
    role_id INTEGER REFERENCES Roles(role_id),
    group_id INTEGER REFERENCES Groups(group_id),
    mail VARCHAR(255),
    login VARCHAR(100),
    password VARCHAR(100),
    name VARCHAR(100),
    surname VARCHAR(100),
    type VARCHAR(100)
);

CREATE TABLE Files (
    file_id SERIAL PRIMARY KEY,
    mongo_file_id TEXT,
    owner_id INTEGER REFERENCES Users(user_id),
    version_id INTEGER,
    type VARCHAR(50),
    name VARCHAR(100),
    full_path VARCHAR(255),
    create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edit_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE FileVersions (
    version_id SERIAL PRIMARY KEY,
    file_id INTEGER,
    user_id INTEGER REFERENCES Users(user_id),
    name VARCHAR(100),
    create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edit_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    mongo_file_id TEXT,
    UNIQUE(file_id, name)
);

ALTER TABLE Files
ADD CONSTRAINT fk_files_version
FOREIGN KEY (version_id)
REFERENCES FileVersions(version_id)
ON DELETE SET NULL;

ALTER TABLE FileVersions
ADD CONSTRAINT fk_fileversions_file
FOREIGN KEY (file_id)
REFERENCES Files(file_id)
ON DELETE CASCADE;

ALTER TABLE Groups 
ADD COLUMN parent_id INTEGER 
REFERENCES Groups(group_id) 
ON DELETE SET NULL;

CREATE TABLE File_Users (
    file_id INTEGER REFERENCES Files(file_id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES Users(user_id) ON DELETE CASCADE,
    access_id INTEGER REFERENCES Access(access_id),
    PRIMARY KEY (file_id, user_id)
);

CREATE TABLE File_Groups (
    file_id INTEGER REFERENCES Files(file_id) ON DELETE CASCADE,
    group_id INTEGER REFERENCES Groups(group_id) ON DELETE CASCADE,
    access_id INTEGER REFERENCES Access(access_id),
    PRIMARY KEY (file_id, group_id)
);

CREATE OR REPLACE FUNCTION get_group_tree(root_id INT)
RETURNS TABLE (
    group_id INT,
    name TEXT,
    description TEXT,
    parent_id INT,
    depth INT,
    path INT[]
)
AS $$
BEGIN
    RETURN QUERY
    WITH RECURSIVE tree AS (
        SELECT
            group_id, name, description, parent_id, 0 AS depth, ARRAY[group_id]
        FROM Groups
        WHERE (root_id IS NULL AND parent_id IS NULL) OR group_id = root_id

        UNION ALL

        SELECT
            g.group_id, g.name, g.description, g.parent_id, t.depth + 1, t.path || g.group_id
        FROM Groups g
        JOIN tree t ON g.parent_id = t.group_id
    )
    SELECT * FROM tree ORDER BY path;
END;
$$ LANGUAGE plpgsql;

INSERT INTO Permissions (name, description) VALUES
('manage_roles', 'Управление ролями'),
('manage_groups', 'Управление группами'),
('manage_users', 'Управление пользователями');

INSERT INTO Access (name) VALUES
('read'),
('write');

INSERT INTO Users (login, password, mail, name, surname, type) 
VALUES ('admin', '$2a$10$FMCEflfMWM0mdyj2laQLmOZ6KbpVH5.I62Hj7wPCzZmYWxYFbCtqG', 'admin@admin.admin', 'admin', 'admin', 'admin')
