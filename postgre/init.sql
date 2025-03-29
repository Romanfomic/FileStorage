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

CREATE TABLE FileVersions (
    version_id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES Users(user_id),
    name VARCHAR(100),
    create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edit_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE Files (
    file_id SERIAL PRIMARY KEY,
    mongo_file_id TEXT,
    group_id INTEGER REFERENCES Groups(group_id),
    owner_id INTEGER REFERENCES Users(user_id),
    access_id INTEGER REFERENCES Access(access_id),
    version_id INTEGER REFERENCES FileVersions(version_id),
    type VARCHAR(50),
    name TEXT,
    full_path TEXT,
    create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edit_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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

INSERT INTO Permissions (name, description) VALUES
('manage_roles', 'Управление ролями'),
('manage_groups', 'Управление группами'),
('manage_users', 'Управление пользователями');

INSERT INTO Access (name) VALUES
('read'),
('write');

INSERT INTO Users (login, password, mail, name, surname, type) 
VALUES ('admin', '$2a$10$FMCEflfMWM0mdyj2laQLmOZ6KbpVH5.I62Hj7wPCzZmYWxYFbCtqG', 'admin@admin.admin', 'admin', 'admin', 'admin')
