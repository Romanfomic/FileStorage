CREATE TABLE Permissions (
    permission_id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    description TEXT
);

CREATE TABLE Roles (
    role_id SERIAL PRIMARY KEY,
    permission_id INTEGER REFERENCES Permissions(permission_id),
    name VARCHAR(100),
    description TEXT
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
    surname VARCHAR(100)
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
    role_id INTEGER REFERENCES Roles(role_id),
    access_id INTEGER REFERENCES Access(access_id),
    version_id INTEGER REFERENCES FileVersions(version_id),
    type VARCHAR(50),
    name TEXT,
    full_path TEXT,
    create_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    edit_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
