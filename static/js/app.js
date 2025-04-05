// Common functionality for the CardMax application

// DOM Content Loaded Event
document.addEventListener('DOMContentLoaded', function() {
    // Initialize the service worker for PWA
    if ('serviceWorker' in navigator) {
        navigator.serviceWorker.register('/static/service-worker.js')
            .then(registration => {
                console.log('Service Worker registered with scope:', registration.scope);
            })
            .catch(error => {
                console.error('Service Worker registration failed:', error);
            });
    }
});

// API Endpoints
const API = {
    cards: '/api/cards',
    card: (id) => `/api/cards/${id}`,
    cardRules: (id) => `/api/cards/${id}/rewards`,
    cardRule: (cardId, ruleId) => `/api/cards/${cardId}/rewards/${ruleId}`,
    recommend: '/api/recommend',
    transactions: '/api/transactions',
    transaction: (id) => `/api/transactions/${id}`
};

// Utility Functions
const Utils = {
    // Format currency
    formatCurrency: (amount) => {
        return new Intl.NumberFormat('en-IN', {
            style: 'currency',
            currency: 'INR',
            minimumFractionDigits: 2
        }).format(amount);
    },
    
    // Format date
    formatDate: (dateString) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-IN', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    },
    
    // Format percentage
    formatPercentage: (value) => {
        return `${value}%`;
    },
    
    // Show error message
    showError: (message) => {
        alert(message);
        console.error(message);
    },
    
    // Show success message
    showSuccess: (message) => {
        alert(message);
    },
    
    // Toggle modal visibility
    toggleModal: (modalId, show) => {
        const modal = document.getElementById(modalId);
        if (show) {
            modal.classList.remove('hidden');
        } else {
            modal.classList.add('hidden');
        }
    },
    
    // Get today's date in YYYY-MM-DD format
    getTodayDate: () => {
        const today = new Date();
        const year = today.getFullYear();
        const month = String(today.getMonth() + 1).padStart(2, '0');
        const day = String(today.getDate()).padStart(2, '0');
        return `${year}-${month}-${day}`;
    },
    
    // Get data from localStorage
    getFromStorage: (key) => {
        const data = localStorage.getItem(key);
        return data ? JSON.parse(data) : null;
    },
    
    // Save data to localStorage
    saveToStorage: (key, data) => {
        localStorage.setItem(key, JSON.stringify(data));
    }
};

// Data Storage (using localStorage for offline capability)
const Storage = {
    // Cards
    getCards: () => {
        return Utils.getFromStorage('cards') || [];
    },
    
    saveCards: (cards) => {
        Utils.saveToStorage('cards', cards);
    },
    
    addCard: (card) => {
        const cards = Storage.getCards();
        // Generate ID if not present
        if (!card.id) {
            card.id = cards.length > 0 ? Math.max(...cards.map(c => c.id)) + 1 : 1;
        }
        cards.push(card);
        Storage.saveCards(cards);
        return card;
    },
    
    updateCard: (card) => {
        const cards = Storage.getCards();
        const index = cards.findIndex(c => c.id === card.id);
        if (index !== -1) {
            cards[index] = card;
            Storage.saveCards(cards);
            return card;
        }
        return null;
    },
    
    deleteCard: (id) => {
        const cards = Storage.getCards();
        const newCards = cards.filter(c => c.id !== id);
        Storage.saveCards(newCards);
        // Also delete related reward rules
        const rules = Storage.getRewardRules();
        const newRules = rules.filter(r => r.cardId !== id);
        Storage.saveRewardRules(newRules);
    },
    
    getCardById: (id) => {
        const cards = Storage.getCards();
        return cards.find(c => c.id === id) || null;
    },
    
    // Reward Rules
    getRewardRules: () => {
        return Utils.getFromStorage('rewardRules') || [];
    },
    
    saveRewardRules: (rules) => {
        Utils.saveToStorage('rewardRules', rules);
    },
    
    getCardRules: (cardId) => {
        const rules = Storage.getRewardRules();
        return rules.filter(r => r.cardId === cardId);
    },
    
    addRewardRule: (rule) => {
        const rules = Storage.getRewardRules();
        // Generate ID if not present
        if (!rule.id) {
            rule.id = rules.length > 0 ? Math.max(...rules.map(r => r.id)) + 1 : 1;
        }
        rules.push(rule);
        Storage.saveRewardRules(rules);
        return rule;
    },
    
    updateRewardRule: (rule) => {
        const rules = Storage.getRewardRules();
        const index = rules.findIndex(r => r.id === rule.id);
        if (index !== -1) {
            rules[index] = rule;
            Storage.saveRewardRules(rules);
            return rule;
        }
        return null;
    },
    
    deleteRewardRule: (id) => {
        const rules = Storage.getRewardRules();
        const newRules = rules.filter(r => r.id !== id);
        Storage.saveRewardRules(newRules);
    },
    
    // Transactions
    getTransactions: () => {
        return Utils.getFromStorage('transactions') || [];
    },
    
    saveTransactions: (transactions) => {
        Utils.saveToStorage('transactions', transactions);
    },
    
    addTransaction: (transaction) => {
        const transactions = Storage.getTransactions();
        // Generate ID if not present
        if (!transaction.id) {
            transaction.id = transactions.length > 0 ? Math.max(...transactions.map(t => t.id)) + 1 : 1;
        }
        transactions.push(transaction);
        Storage.saveTransactions(transactions);
        return transaction;
    },
    
    updateTransaction: (transaction) => {
        const transactions = Storage.getTransactions();
        const index = transactions.findIndex(t => t.id === transaction.id);
        if (index !== -1) {
            transactions[index] = transaction;
            Storage.saveTransactions(transactions);
            return transaction;
        }
        return null;
    },
    
    deleteTransaction: (id) => {
        const transactions = Storage.getTransactions();
        const newTransactions = transactions.filter(t => t.id !== id);
        Storage.saveTransactions(newTransactions);
    },
    
    getTransactionById: (id) => {
        const transactions = Storage.getTransactions();
        return transactions.find(t => t.id === id) || null;
    },
    
    // Filtered transactions
    getFilteredTransactions: (filters) => {
        let transactions = Storage.getTransactions();
        
        if (filters.dateFrom) {
            transactions = transactions.filter(t => new Date(t.date) >= new Date(filters.dateFrom));
        }
        
        if (filters.dateTo) {
            transactions = transactions.filter(t => new Date(t.date) <= new Date(filters.dateTo));
        }
        
        if (filters.category && filters.category !== '') {
            transactions = transactions.filter(t => t.category === filters.category);
        }
        
        if (filters.cardId && filters.cardId !== '') {
            transactions = transactions.filter(t => t.cardId === parseInt(filters.cardId));
        }
        
        return transactions;
    }
};

// Recommendation Engine
const RecommendationEngine = {
    // Get card recommendation based on merchant, category, and amount
    getRecommendation: (merchant, category, amount) => {
        const cards = Storage.getCards();
        const results = [];
        
        // If no cards, return empty results
        if (cards.length === 0) {
            return results;
        }
        
        // Calculate rewards for each card
        for (const card of cards) {
            const rules = Storage.getCardRules(card.id);
            let bestRule = null;
            let bestRewardRate = card.defaultRewardRate;
            let rewardType = 'Points';
            let pointValue = 1;
            
            // Find the best matching rule
            for (const rule of rules) {
                if (
                    (rule.type === 'Merchant' && rule.entityName.toLowerCase() === merchant.toLowerCase()) ||
                    (rule.type === 'Category' && rule.entityName.toLowerCase() === category.toLowerCase())
                ) {
                    if (rule.rewardRate > bestRewardRate) {
                        bestRule = rule;
                        bestRewardRate = rule.rewardRate;
                        rewardType = rule.rewardType;
                        pointValue = rule.pointValue || 1;
                    }
                }
            }
            
            // Calculate reward value
            let rewardAmount = (amount * bestRewardRate) / 100;
            let cashValue = rewardAmount;
            
            if (rewardType === 'Points' || rewardType === 'Miles') {
                cashValue = rewardAmount * pointValue;
            }
            
            results.push({
                card: card,
                rewardRate: bestRewardRate,
                rewardType: rewardType,
                rewardAmount: rewardAmount,
                cashValue: cashValue,
                rule: bestRule
            });
        }
        
        // Sort by cash value (highest first)
        results.sort((a, b) => b.cashValue - a.cashValue);
        
        return results;
    }
};
