DROP INDEX idx_unique_reward_rule;

-- Drop the reward rules table first (due to foreign key constraint)
DROP TABLE IF EXISTS predefined_reward_rules;

-- Drop the predefined cards table
DROP TABLE IF EXISTS predefined_cards;