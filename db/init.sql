USE collab_doc;

-- 创建用户表
CREATE TABLE users (
                       id INT PRIMARY KEY AUTO_INCREMENT,
                       username VARCHAR(50) UNIQUE NOT NULL,
                       password VARCHAR(255) NOT NULL,
                       email VARCHAR(100),
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建文档表
CREATE TABLE documents (
                           id INT PRIMARY KEY AUTO_INCREMENT,
                           title VARCHAR(255) NOT NULL,
                           content LONGTEXT,
                           owner_id INT,
                           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                           updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                           FOREIGN KEY (owner_id) REFERENCES users(id)
);

-- 修改数据库编码
ALTER DATABASE collab_doc CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 修改用户表编码
ALTER TABLE users CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 修改文档表编码
ALTER TABLE documents CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE doc_versions (
                              id INT PRIMARY KEY AUTO_INCREMENT,
                              doc_id INT NOT NULL COMMENT '关联的文档ID',
                              content LONGTEXT COMMENT '该版本的内容',
                              version_name VARCHAR(100) COMMENT '版本备注（如：第一次修改）',
                              created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                              FOREIGN KEY (doc_id) REFERENCES documents(id) ON DELETE CASCADE
) COMMENT='文档版本历史表';

-- 完善用户表：增加手机号、头像、角色、个人简介等
ALTER TABLE users
    ADD COLUMN phone VARCHAR(20) AFTER email,
    ADD COLUMN avatar VARCHAR(255) DEFAULT 'default_avatar.png' AFTER phone,
    ADD COLUMN bio TEXT AFTER avatar,
    ADD COLUMN role VARCHAR(20) DEFAULT 'editor' COMMENT 'admin, editor, viewer';

-- 创建操作日志表
CREATE TABLE operation_logs (
                                id INT PRIMARY KEY AUTO_INCREMENT,
                                user_id INT,
                                action VARCHAR(255) COMMENT '操作内容',
                                doc_id INT,
                                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                FOREIGN KEY (user_id) REFERENCES users(id)
);

-- 增加文档分类和标签
ALTER TABLE documents ADD COLUMN category VARCHAR(50) DEFAULT '默认';
ALTER TABLE documents ADD COLUMN tags VARCHAR(255) DEFAULT ''; -- 存储如 "作业,重要"

-- 评论表 (3.2)
CREATE TABLE comments (
                          id INT PRIMARY KEY AUTO_INCREMENT,
                          doc_id INT,
                          user_id INT,
                          content TEXT,
                          line_num INT DEFAULT 0, -- 实现行内评论
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          FOREIGN KEY (doc_id) REFERENCES documents(id) ON DELETE CASCADE,
                          FOREIGN KEY (user_id) REFERENCES users(id)
);

-- 聊天记录表 (4.2)
CREATE TABLE chat_messages (
                               id INT PRIMARY KEY AUTO_INCREMENT,
                               doc_id INT,
                               user_id INT,
                               username VARCHAR(50),
                               content TEXT,
                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建文件夹表
CREATE TABLE folders (
                         id INT PRIMARY KEY AUTO_INCREMENT,
                         name VARCHAR(50) NOT NULL,
                         user_id INT NOT NULL, -- 区分不同用户的文件夹
                         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 修改文档表：增加“收藏”状态和“文件夹关联”
ALTER TABLE documents
    ADD COLUMN is_starred BOOLEAN DEFAULT FALSE COMMENT '是否为重要记录',
    ADD COLUMN folder_id INT DEFAULT 0 COMMENT '所属文件夹ID，0代表根目录';


create table chat_messages
(
    id         int auto_increment
        primary key,
    doc_id     int                                 null,
    user_id    int                                 null,
    username   varchar(50)                         null,
    content    text                                null,
    created_at timestamp default CURRENT_TIMESTAMP null
);

create table folders
(
    id         int auto_increment
        primary key,
    name       varchar(50)                         not null,
    user_id    int                                 not null,
    created_at timestamp default CURRENT_TIMESTAMP null
);

create table users
(
    id         int auto_increment
        primary key,
    username   varchar(50)                               not null,
    password   varchar(255)                              not null,
    email      varchar(100)                              null,
    phone      varchar(20)                               null,
    avatar     varchar(255) default 'default_avatar.png' null,
    bio        text                                      null,
    created_at timestamp    default CURRENT_TIMESTAMP    null,
    role       varchar(20)  default 'editor'             null comment 'admin, editor, viewer',
    constraint username
        unique (username)
);

create table documents
(
    id         int auto_increment
        primary key,
    title      varchar(255)                           not null,
    content    longtext                               null,
    owner_id   int                                    null,
    created_at timestamp    default CURRENT_TIMESTAMP null,
    updated_at timestamp    default CURRENT_TIMESTAMP null on update CURRENT_TIMESTAMP,
    category   varchar(50)  default '默认'            null,
    tags       varchar(255) default ''                null,
    is_starred tinyint(1)   default 0                 null comment '是否为重要记录',
    folder_id  int          default 0                 null comment '所属文件夹ID，0代表根目录',
    constraint documents_ibfk_1
        foreign key (owner_id) references users (id)
);

create table comments
(
    id         int auto_increment
        primary key,
    doc_id     int                                 null,
    user_id    int                                 null,
    content    text                                null,
    line_num   int       default 0                 null,
    created_at timestamp default CURRENT_TIMESTAMP null,
    constraint comments_ibfk_1
        foreign key (doc_id) references documents (id)
            on delete cascade,
    constraint comments_ibfk_2
        foreign key (user_id) references users (id)
);

create index doc_id
    on comments (doc_id);

create index user_id
    on comments (user_id);

create table doc_versions
(
    id           int auto_increment
        primary key,
    doc_id       int                                 not null comment '关联的文档ID',
    content      longtext                            null comment '该版本的内容',
    version_name varchar(100)                        null comment '版本备注（如：第一次修改）',
    created_at   timestamp default CURRENT_TIMESTAMP null,
    constraint doc_versions_ibfk_1
        foreign key (doc_id) references documents (id)
            on delete cascade
)
    comment '文档版本历史表';

create index doc_id
    on doc_versions (doc_id);

create index owner_id
    on documents (owner_id);

create table operation_logs
(
    id         int auto_increment
        primary key,
    user_id    int                                 null,
    action     varchar(255)                        null comment '操作内容',
    doc_id     int                                 null,
    created_at timestamp default CURRENT_TIMESTAMP null,
    constraint operation_logs_ibfk_1
        foreign key (user_id) references users (id)
);

create index user_id
    on operation_logs (user_id);

