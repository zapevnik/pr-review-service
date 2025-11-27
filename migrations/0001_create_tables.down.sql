DROP INDEX IF EXISTS idx_pr_reviewers_reviewer;
DROP INDEX IF EXISTS idx_pull_requests_status;
DROP INDEX IF EXISTS idx_pull_requests_author;
DROP INDEX IF EXISTS idx_users_active_team;
DROP INDEX IF EXISTS idx_users_team;

DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
