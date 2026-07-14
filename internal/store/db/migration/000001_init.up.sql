CREATE TABLE "Posts" (
  "id" integer PRIMARY KEY,
  "title" varchar NOT NULL,
  "description" varchar NOT NULL,
  "author" varchar NOT NULL
);

CREATE TABLE "Comments" (
  "id" integer PRIMARY KEY,
  "comment" varchar NOT NULL,
  "post_id" integer,
  "author" varchar NOT NULL
);

ALTER TABLE "Comments" ADD FOREIGN KEY ("post_id") REFERENCES "Posts" ("id") DEFERRABLE INITIALLY IMMEDIATE;
