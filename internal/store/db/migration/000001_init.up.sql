CREATE TABLE categories (
  "id" integer PRIMARY KEY,
  "name" varchar NOT NULL UNIQUE
);
CREATE TABLE posts (
  "id" integer PRIMARY KEY,
  "title" varchar NOT NULL,
  "description" varchar NOT NULL,
  "author" varchar NOT NULL,
  "category_id" integer NOT NULL
);

CREATE TABLE comments (
  "id" integer PRIMARY KEY,
  "comment" varchar NOT NULL,
  "post_id" integer,
  "author" varchar NOT NULL
);

ALTER TABLE comments ADD FOREIGN KEY ("post_id") REFERENCES posts ("id") DEFERRABLE INITIALLY IMMEDIATE;
ALTER TABLE posts ADD FOREIGN KEY ("category_id") REFERENCES categories ("id") DEFERRABLE INITIALLY IMMEDIATE;
