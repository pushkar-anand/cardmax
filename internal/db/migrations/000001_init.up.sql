CREATE TABLE cards
(
    -- ID: Unique identifier for each card record.
    -- Automatically increments with each new entry.
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Name: The name given to the credit card (e.g., 'Platinum Rewards').
    -- Cannot be empty.
    name                TEXT NOT NULL,

    -- Issuer: The financial institution that issued the card (e.g., 'Bank of Example').
    -- Cannot be empty.
    issuer              TEXT NOT NULL,

    -- Last4Digits: The last four digits of the credit card number.
    -- Stored as text to preserve leading zeros if any.
    -- It Must be exactly 4 characters long. Cannot be empty.
    last4_digits        TEXT NOT NULL CHECK (length(last4_digits) = 4),

    -- ExpiryDate: The expiration date of the card stored as 'YYYY-MM'.
    -- Stored as text for flexibility. Cannot be empty.
    expiry_date         TEXT NOT NULL,

    -- DefaultRewardRate: The standard reward rate associated with the card (e.g., 1.5 for 1.5%).
    -- Stored as a real number to allow for decimal values. It defaults to 0.0 if not specified.
    default_reward_rate REAL DEFAULT 0.0,

    -- CardType: The type or network of the card (e.g., 'Visa', 'Mastercard', 'Amex').
    -- Cannot be empty.
    card_type           TEXT NOT NULL
);