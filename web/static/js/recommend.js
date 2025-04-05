// Recommendation functionality

document.addEventListener('DOMContentLoaded', function() {
    // DOM Elements
    const recommendForm = document.getElementById('recommend-form');
    const recommendationResults = document.getElementById('recommendation-results');
    const bestCardResult = document.getElementById('best-card-result');
    const otherCardsResults = document.getElementById('other-cards-results');
    const saveTransactionBtn = document.getElementById('save-transaction-btn');
    
    const transactionModal = document.getElementById('transaction-modal');
    const saveTransactionForm = document.getElementById('save-transaction-form');
    const closeTransactionModal = document.querySelector('.close-transaction-modal');
    const cancelTransactionBtn = document.getElementById('cancel-transaction');
    const transactionCardSelect = document.getElementById('transaction-card');
    
    // Current recommendation results
    let currentRecommendation = null;
    
    // Initialize
    initializeForm();
    
    // Event Listeners
    recommendForm.addEventListener('submit', getRecommendation);
    saveTransactionBtn.addEventListener('click', showSaveTransactionForm);
    closeTransactionModal.addEventListener('click', hideTransactionForm);
    cancelTransactionBtn.addEventListener('click', hideTransactionForm);
    saveTransactionForm.addEventListener('submit', saveTransaction);
    
    // Functions
    function initializeForm() {
        // Set today's date as default
        const today = Utils.getTodayDate();
        if (document.getElementById('transaction-date')) {
            document.getElementById('transaction-date').value = today;
        }
    }
    
    function getRecommendation(e) {
        e.preventDefault();
        
        const merchant = document.getElementById('merchant').value.toLowerCase();
        const category = document.getElementById('category').value.toLowerCase();
        const amount = parseFloat(document.getElementById('amount').value);
        
        if (!category) {
            Utils.showError('Please select a category');
            return;
        }
        
        if (isNaN(amount) || amount <= 0) {
            Utils.showError('Please enter a valid amount');
            return;
        }
        
        // Show loading state
        recommendationResults.innerHTML = '<div class="loading">Calculating best cards...</div>';
        recommendationResults.classList.remove('hidden');
        
        // API request for recommendation
        fetch('/api/recommend', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                merchant: merchant,
                category: category,
                amount: amount
            }),
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Error getting recommendations');
            }
            return response.json();
        })
        .then(data => {
            // Store for later use
            currentRecommendation = {
                merchant,
                category,
                amount,
                results: data.all_cards
            };
            
            // Display the recommendation
            displayRecommendationFromAPI(data, amount);
        })
        .catch(error => {
            console.error('Error:', error);
            recommendationResults.innerHTML = `
                <div class="error-message">
                    <p>Sorry, we couldn't get recommendations right now. Please try again.</p>
                </div>
            `;
        });
    }
    }
    
    function displayRecommendationFromAPI(data, amount) {
        if (!data.best_card) {
            bestCardResult.innerHTML = `
                <p>No cards found. <a href="/cards">Add your first card</a> to get recommendations.</p>
            `;
            otherCardsResults.innerHTML = '';
            saveTransactionBtn.disabled = true;
            return;
        }
        
        // Best card
        const bestCard = data.best_card;
        bestCardResult.innerHTML = createCardResultHTMLFromAPI(bestCard, amount, true);
        
        // Other cards
        let otherCardsHTML = '';
        if (data.all_cards && data.all_cards.length > 1) {
            otherCardsHTML = '<div class="other-cards-accordion">';
            for (let i = 1; i < data.all_cards.length; i++) {
                otherCardsHTML += createCardResultHTMLFromAPI(data.all_cards[i], amount, false);
            }
            otherCardsHTML += '</div>';
        } else {
            otherCardsHTML = '<p>No other cards available.</p>';
        }
        otherCardsResults.innerHTML = otherCardsHTML;
        
        // Show results section
        recommendationResults.classList.remove('hidden');
        
        // Enable save transaction button
        saveTransactionBtn.disabled = false;
        
        // Add event listeners for accordion
        document.querySelectorAll('.card-result-header').forEach(header => {
            if (!header.closest('.best-card')) {
                header.addEventListener('click', function() {
                    const card = this.closest('.card-result-item');
                    card.classList.toggle('expanded');
                });
            }
        });
    }
    
    function createCardResultHTMLFromAPI(result, amount, isBest) {
        const card = result.card;
        let rewardDisplay = '';
        
        if (result.reward_type === 'Cashback') {
            rewardDisplay = `${Utils.formatCurrency(result.cash_value)} cashback (${result.reward_rate}%)`;
        } else {
            rewardDisplay = `${result.reward_value.toFixed(0)} ${result.reward_type} (${result.reward_rate}%)`;
            rewardDisplay += ` worth ${Utils.formatCurrency(result.cash_value)}`;
        }
        
        // Extract last4 digits from the card.Name if available, otherwise use card_key
        const cardIdentifier = card.last4_digits ? `*${card.last4_digits}` : card.card_key;
        
        return `
            <div class="card-result-item ${isBest ? 'best-card-item' : ''}" data-key="${card.card_key}">
                <div class="card-result-header">
                    <div class="card-result-name">${card.name} (${cardIdentifier})</div>
                    <div class="card-result-reward">${rewardDisplay}</div>
                    ${!isBest ? '<div class="expand-icon">▼</div>' : ''}
                </div>
                <div class="card-result-details ${!isBest ? 'collapsed' : ''}">
                    <div>On ${Utils.formatCurrency(amount)} purchase</div>
                    ${result.rule ? `<div>Special rate for ${result.rule.type}: ${result.rule.entity_name}</div>` : ''}
                    <div class="card-issuer">Issued by: ${card.issuer}</div>
                </div>
            </div>
        `;
    }
    
    function displayRecommendation(results, amount) {
        // Legacy function kept for compatibility
        if (results.length === 0) {
            bestCardResult.innerHTML = `
                <p>No cards found. <a href="/cards">Add your first card</a> to get recommendations.</p>
            `;
            otherCardsResults.innerHTML = '';
        } else {
            // Best card (first in the sorted results)
            const bestCard = results[0];
            bestCardResult.innerHTML = createCardResultHTML(bestCard, amount);
            
            // Other cards
            let otherCardsHTML = '';
            if (results.length > 1) {
                for (let i = 1; i < results.length; i++) {
                    otherCardsHTML += createCardResultHTML(results[i], amount);
                }
            } else {
                otherCardsHTML = '<p>No other cards available.</p>';
            }
            otherCardsResults.innerHTML = otherCardsHTML;
        }
        
        // Show results section
        recommendationResults.classList.remove('hidden');
        
        // Enable/disable save transaction button
        saveTransactionBtn.disabled = results.length === 0;
    }
    
    function createCardResultHTML(result, amount) {
        const card = result.card;
        let rewardDisplay = '';
        
        if (result.rewardType === 'Cashback') {
            rewardDisplay = `${Utils.formatCurrency(result.cashValue)} cashback (${result.rewardRate}%)`;
        } else {
            rewardDisplay = `${result.rewardAmount.toFixed(0)} ${result.rewardType} (${result.rewardRate}%)`;
            if (result.pointValue) {
                rewardDisplay += ` worth ${Utils.formatCurrency(result.cashValue)}`;
            }
        }
        
        return `
            <div class="card-result-item" data-id="${card.id}">
                <div class="card-result-header">
                    <div class="card-result-name">${card.name} (*${card.last4Digits})</div>
                    <div class="card-result-reward">${rewardDisplay}</div>
                </div>
                <div class="card-result-details">
                    <div>On ${Utils.formatCurrency(amount)} purchase</div>
                    ${result.rule ? `<div>Special rate for ${result.rule.type}: ${result.rule.entityName}</div>` : ''}
                </div>
            </div>
        `;
    }
    
    function showSaveTransactionForm() {
        if (!currentRecommendation || currentRecommendation.results.length === 0) {
            return;
        }
        
        // Populate form with current recommendation data
        const form = saveTransactionForm;
        form.elements['date'].value = Utils.getTodayDate();
        form.elements['merchantName'].value = currentRecommendation.merchant;
        form.elements['category'].value = currentRecommendation.category;
        form.elements['amount'].value = currentRecommendation.amount;
        
        // Populate card select
        populateCardSelect();
        
        // Set default card to best recommendation (API format)
        if (currentRecommendation.results[0].card && currentRecommendation.results[0].card.card_key) {
            // Find the user's card that matches the best recommendation card key
            const cards = Storage.getCards();
            const bestCardKey = currentRecommendation.results[0].card.card_key;
            const userCard = cards.find(c => c.cardKey === bestCardKey);
            
            if (userCard) {
                form.elements['cardId'].value = userCard.id;
            }
        }
        
        // Set reward earned based on selected card
        updateRewardEarned();
        
        Utils.toggleModal('transaction-modal', true);
    }
    
    function hideTransactionForm() {
        Utils.toggleModal('transaction-modal', false);
    }
    
    function populateCardSelect() {
        const cards = Storage.getCards();
        let options = '';
        
        cards.forEach(card => {
            options += `<option value="${card.id}">${card.name} (*${card.last4Digits})</option>`;
        });
        
        transactionCardSelect.innerHTML = options;
        
        // Add change event to update reward earned
        transactionCardSelect.addEventListener('change', updateRewardEarned);
    }
    
    function updateRewardEarned() {
        if (!currentRecommendation) return;
        
        const selectedCardId = parseInt(transactionCardSelect.value);
        
        // For API response format
        if (currentRecommendation.results[0] && currentRecommendation.results[0].card && currentRecommendation.results[0].card.card_key) {
            // We need to find the result that matches the selected user card
            const userCards = Storage.getCards();
            const selectedUserCard = userCards.find(c => c.id === selectedCardId);
            
            if (selectedUserCard) {
                // Find the card in the API results that matches the selected user card's key
                const result = currentRecommendation.results.find(r => 
                    r.card.card_key === selectedUserCard.cardKey);
                
                if (result) {
                    saveTransactionForm.elements['rewardEarned'].value = result.cash_value.toFixed(2);
                    return;
                }
            }
        } else {
            // Legacy format (client-side calculation)
            const result = currentRecommendation.results.find(r => r.card.id === selectedCardId);
            
            if (result) {
                saveTransactionForm.elements['rewardEarned'].value = result.cashValue.toFixed(2);
                return;
            }
        }
        
        // Default fallback if no match found
        saveTransactionForm.elements['rewardEarned'].value = '0.00';
    }
    
    function saveTransaction(e) {
        e.preventDefault();
        
        const form = saveTransactionForm;
        
        const transaction = {
            date: form.elements['date'].value,
            merchantName: form.elements['merchantName'].value,
            category: form.elements['category'].value,
            amount: parseFloat(form.elements['amount'].value),
            cardId: parseInt(form.elements['cardId'].value),
            rewardEarned: parseFloat(form.elements['rewardEarned'].value),
            notes: form.elements['notes'].value
        };
        
        Storage.addTransaction(transaction);
        hideTransactionForm();
        
        Utils.showSuccess('Transaction saved successfully!');
    }
});
