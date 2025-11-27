CREATE TABLE teams (
  team_name TEXT PRIMARY KEY
);

CREATE TABLE users (
  user_id TEXT PRIMARY KEY,                                       
  team_name TEXT NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
  user_name TEXT NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE pull_requests (
  pr_id TEXT PRIMARY KEY,                                       
  pr_title TEXT NOT NULL,
  author_id TEXT NOT NULL REFERENCES users(user_id),              
  pr_status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  merged_at TIMESTAMPTZ,
  pr_version BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE pr_reviewers (
  pr_id TEXT NOT NULL REFERENCES pull_requests(pr_id) ON DELETE CASCADE,  
  reviewer_id TEXT NOT NULL REFERENCES users(user_id),                    
  PRIMARY KEY (pr_id, reviewer_id)
);


CREATE INDEX idx_users_team ON users(team_name);
CREATE INDEX idx_users_active_team ON users(team_name, is_active);

CREATE INDEX idx_pull_requests_author ON pull_requests(author_id);
CREATE INDEX idx_pull_requests_status ON pull_requests(pr_status);

CREATE INDEX idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);
