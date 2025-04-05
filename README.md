# CardMax

A credit card management and recommendation application.

## Features

- Manage your credit cards
- Get recommendations for the best card to use for a purchase
- Track your transactions
- View rewards and benefits

## API Documentation

### Predefined Cards

#### Get All Predefined Cards

```
GET /api/predefined-cards
```

Returns a list of all predefined credit cards in the system.

Example response:
```json
[
  {
    "card_key": "ICICI-APAY",
    "name": "ICICI Amazon Pay Credit Card",
    "issuer": "ICICI",
    "card_type": "Visa",
    "default_reward_rate": 1.0,
    "reward_type": "Cashback",
    "point_value": 1.0,
    "annual_fee": 0,
    "annual_fee_waiver": "",
    "reward_rules": [
      {
        "type": "Merchant",
        "entity_name": "amazon",
        "reward_rate": 5.0,
        "reward_type": "Cashback"
      }
    ],
    "benefits": [
      "5% unlimited cashback on Amazon.in shopping (using Amazon Prime)"
    ]
  }
]
```

#### Get Predefined Card by Key

```
GET /api/predefined-cards/{key}
```

Returns a specific predefined credit card by its key.

Example response:
```json
{
  "card_key": "ICICI-APAY",
  "name": "ICICI Amazon Pay Credit Card",
  "issuer": "ICICI",
  "card_type": "Visa",
  "default_reward_rate": 1.0,
  "reward_type": "Cashback",
  "point_value": 1.0,
  "annual_fee": 0,
  "annual_fee_waiver": "",
  "reward_rules": [
    {
      "type": "Merchant",
      "entity_name": "amazon",
      "reward_rate": 5.0,
      "reward_type": "Cashback"
    }
  ],
  "benefits": [
    "5% unlimited cashback on Amazon.in shopping (using Amazon Prime)"
  ]
}
```