// Homepage functionality

document.addEventListener('DOMContentLoaded', function() {
    // Initialize the quick recommendation form
    if (RecommendationUI && typeof RecommendationUI.initializeRecommendationForm === 'function') {
        RecommendationUI.initializeRecommendationForm('quick-recommend-form', 'recommendation-result', true);
    } else {
        console.error('RecommendationUI not found. Make sure recommend.js is loaded before index.js');
    }
    
    // Display recent transactions
    displayRecentTransactions();
    
    // Display user's cards
    displayUserCards();
    
    // Functions
    function displayRecentTransactions() {
        const transactionsList = document.querySelector('.transactions-list');
        if (!transactionsList) return;
        
        const transactions = Storage.getTransactions();
        
        if (transactions.length === 0) {
            transactionsList.innerHTML = '<p>No transactions recorded yet.</p>';
            return;
        }
        
        // Sort by date (newest first)
        transactions.sort((a, b) => new Date(b.date) - new Date(a.date));
        
        // Take only the 5 most recent
        const recentTransactions = transactions.slice(0, 5);
        
        let html = '<ul class="transaction-cards">';
        
        for (const transaction of recentTransactions) {
            const card = Storage.getCardById(transaction.cardId);
            const cardName = card ? card.name : 'Unknown Card';
            
            html += `
                <li class="transaction-card">
                    <div class="transaction-date">${Utils.formatDate(transaction.date)}</div>
                    <div class="transaction-details">
                        <div class="transaction-merchant">${transaction.merchantName}</div>
                        <div class="transaction-category">${transaction.category}</div>
                        <div class="transaction-amount">${Utils.formatCurrency(transaction.amount)}</div>
                    </div>
                    <div class="transaction-card-used">${cardName}</div>
                    <div class="transaction-reward">+${Utils.formatCurrency(transaction.rewardEarned)}</div>
                </li>
            `;
        }
        
        html += '</ul>';
        transactionsList.innerHTML = html;
    }
    
    function displayUserCards() {
        const cardsGrid = document.querySelector('.cards-grid');
        if (!cardsGrid) return;
        
        const cards = Storage.getCards();
        
        if (cards.length === 0) {
            cardsGrid.innerHTML = '<p>No cards added yet. <a href="/cards">Add your first card</a>.</p>';
            return;
        }
        
        let html = '';
        
        for (const card of cards) {
            html += `
                <div class="card-preview">
                    <div class="card-header">
                        <div class="card-name">${card.name}</div>
                        <div class="card-issuer">${card.issuer}</div>
                    </div>
                    <div class="card-body">
                        <div class="card-number">**** **** **** ${card.last4Digits}</div>
                        <div class="card-expiry">Expires: ${card.expiryDate || 'N/A'}</div>
                    </div>
                    <div class="card-reward">
                        <div class="reward-rate">${card.defaultRewardRate}% ${card.rewardType}</div>
                    </div>
                </div>
            `;
        }
        
        cardsGrid.innerHTML = html;
    }
});