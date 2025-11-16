
CREATE TABLE pull_requests (
  pull_request_id VARCHAR PRIMARY KEY, -- внешний id
  pull_request_name TEXT NOT NULL,
  author_id VARCHAR NOT NULL,          -- FK -> users.user_id
  status VARCHAR(10) NOT NULL,         -- 'OPEN' | 'MERGED'
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  merged_at TIMESTAMP WITH TIME ZONE NULL
);


CREATE TABLE pr_reviewers (
  pull_request_id VARCHAR REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
  user_id VARCHAR REFERENCES users(user_id),
  PRIMARY KEY (pull_request_id, user_id)
);
-- ограничение: максимум 2 строки на pull_request_id (можно контролировать в коде)


CREATE UNIQUE INDEX ux_pull_request_id ON pull_requests(pull_request_id);



