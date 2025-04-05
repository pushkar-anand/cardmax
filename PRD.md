# Product Requirements Document: CardMax

## 1. Executive Summary

CardMax is a mobile-first web application with PWA support designed to help users maximize credit card rewards by recommending the optimal card for each purchase. The app analyzes the user's credit card portfolio and provides real-time recommendations to ensure users get the highest possible return on their spending.

## 2. Technology Stack

### Frontend
- HTML5
- CSS3 (with responsive design principles)
- Vanilla JavaScript
- Progressive Web App (PWA) features

### Backend
- Golang
- Golang HTML templates for server-side rendering
- RESTful API endpoints for data operations
- SQLite or PostgreSQL for data storage

## 3. Phase 1 Core Features

### 3.1 Card Management
- Add, edit, and delete credit cards
- Required card details:
    - Card name (e.g., "HDFC Diners")
    - Card issuer (e.g., HDFC, ICICI, Axis, AMEX)
    - Last 4 digits (for identification)
    - Expiration date
    - Card type (Visa, Mastercard, AMEX, etc.)
    - Default reward rate (e.g., 1% on all purchases)

### 3.2 Reward Structure Management
- Predefined reward structures for the user's specific cards:
    - Standard points/cashback rates
    - Category-specific multipliers (e.g., 5x on dining)
    - Merchant-specific rewards (e.g., 5% on Amazon)
    - Point valuation (e.g., 1 point = ₹0.25)
- Admin interface to update reward structures when banks change terms

### 3.3 Purchase Recommendation Engine
- Input fields:
    - Merchant name or selection from common merchants
    - Purchase category (dropdown)
    - Estimated purchase amount
- Output:
    - Ranked list of cards with expected rewards value
    - Clear indication of best card to use
    - Expected rewards calculation (points, cashback, or value)

### 3.4 Transaction Logging
- Manual entry of transactions:
    - Date
    - Merchant
    - Category
    - Amount
    - Card used
    - Reward earned
- Transaction history view with basic filtering
- Simple summary of reward earnings

## 4. User Interface Requirements

### 4.1 Responsive Design
- Mobile-first approach
- Support for various screen sizes (mobile, tablet, desktop)
- Touch-friendly interface elements
- Minimum viable screen width: 320px

### 4.2 Key Screens
- Home/Dashboard:
    - Quick recommendation widget
    - Summary of recent transactions
    - Card overview
- Card Management:
    - List view of all cards
    - Add/edit card form
- Purchase Recommendation:
    - Simple form for entering purchase details
    - Clear visual hierarchy for recommendations
- Transaction History:
    - Chronological list with basic filters
    - Simple statistics on rewards earned

### 4.3 PWA Features
- Installable on home screen
- Offline functionality for core features
- App manifest and icons
- Service worker for caching

## 5. Data Requirements

### 5.1 Data Models

#### Card
- ID (unique identifier)
- Name
- Issuer
- Last4Digits
- ExpiryDate
- DefaultRewardRate
- CardType

#### RewardRule
- ID
- CardID (foreign key)
- Type (Category, Merchant)
- EntityName (category name or merchant name)
- RewardRate
- RewardType (Points, Cashback, Miles)
- PointValue (monetary value per point)

#### Transaction
- ID
- Date
- MerchantName
- Category
- Amount
- CardID (foreign key)
- RewardEarned
- Notes

### 5.2 Data Storage
- LocalStorage for offline capabilities
- Server-side database for persistence
- Secure handling of card information

## 6. API Endpoints

### 6.1 Cards API
- GET /api/cards - List all cards
- POST /api/cards - Create new card
- GET /api/cards/{id} - Get card details
- PUT /api/cards/{id} - Update card
- DELETE /api/cards/{id} - Delete card

### 6.2 Rewards API
- GET /api/cards/{id}/rewards - Get reward rules for a card
- POST /api/cards/{id}/rewards - Add reward rule
- PUT /api/cards/{id}/rewards/{ruleId} - Update reward rule
- DELETE /api/cards/{id}/rewards/{ruleId} - Delete reward rule

### 6.3 Recommendation API
- POST /api/recommend - Get card recommendation
    - Request body: merchant, category, amount
    - Response: Ranked list of cards with reward calculations

### 6.4 Transactions API
- GET /api/transactions - List transactions
- POST /api/transactions - Record new transaction
- GET /api/transactions/{id} - Get transaction details
- PUT /api/transactions/{id} - Update transaction
- DELETE /api/transactions/{id} - Delete transaction

## 7. Security Requirements

### 7.1 Data Protection
- HTTPS for all communications
- Hashing of sensitive data
- No storage of complete card numbers
- No storage of CVV codes

### 7.2 Authentication
- Simple username/password authentication
- Session management
- CSRF protection

## 8. Development Guidelines

### 8.1 Frontend
- Mobile-first CSS with media queries
- Semantic HTML5 elements
- JavaScript ES6+ features
- Modular JavaScript organization
- Service worker for offline functionality

### 8.2 Backend
- RESTful API design principles
- Golang standard project layout
- HTML templates for server-side rendering
- JSON for API responses
- Proper error handling and logging

## 9. Performance Requirements

- Page load time < 2 seconds on 4G connection
- Time to interactive < 3 seconds
- Recommendation calculation < 1 second
- Smooth scrolling and transitions
- Minimal data usage

## 10. Testing Requirements

- Cross-browser testing (Chrome, Safari, Firefox)
- Responsive design testing on various screen sizes
- Offline functionality testing
- API endpoint testing
- Data validation testing

## 11. Deployment Considerations

- Simple hosting setup (could use DigitalOcean, Heroku, etc.)
- CI/CD pipeline for automated testing and deployment
- Database backup strategy
- HTTPS certificate setup
- Basic monitoring and logging

## 12. Future Considerations (Phase 2+)

- Milestone tracking for spending goals
- Expanded card database beyond user's current cards
- Advanced analytics and reporting
- Automatic transaction import
- Push notifications for recommendations