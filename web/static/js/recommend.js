// Recommendation functionality
const RecommendationUI = {
    // Get recommendation from the API
    getRecommendationFromAPI: function(merchant, category, amount, userCards, onSuccess, onError) {
        console.log('Sending request with:', {merchant, category, amount, userCards});
        
        fetch('/api/recommend', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                merchant: merchant,
                category: category,
                amount: amount,
                user_cards: userCards
            }),
        })
        .then(response => {
            console.log('Response status:', response.status);
            if (!response.ok) {
                throw new Error(`Error getting recommendations: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            console.log('Received data:', data);
            
            // Store for later use
            const currentRecommendation = {
                merchant,
                category,
                amount,
                results: data.all_cards || []
            };
            
            // Call success callback with results
            if (onSuccess && typeof onSuccess === 'function') {
                onSuccess(data, currentRecommendation);
            }
        })
        .catch(error => {
            console.error('Error:', error);
            
            // Call error callback
            if (onError && typeof onError === 'function') {
                onError(error);
            }
        });
    },
    
    // Create HTML for a card result item
    createCardResultHTML: function(result, amount, isBest) {
        try {
            console.log('Creating HTML for card result:', result);
            
            if (!result || !result.card) {
                console.error('Invalid result object:', result);
                return '<div class="error">Error: Invalid card data</div>';
            }
            
            const card = result.card;
            let rewardDisplay = '';
            
            try {
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
            } catch (error) {
                console.error('Error processing card result:', error, result);
                return `<div class="error">Error processing card: ${error.message}</div>`;
            }
        } catch (error) {
            console.error('Error in createCardResultHTML:', error);
            return `<div class="error">Error: ${error.message}</div>`;
        }
    },
    
    // Display a recommendation in a container
    displayRecommendation: function(container, data, amount, showAllCards = true) {
        try {
            console.log('Displaying recommendation in container:', container, data);
            
            if (!data || !data.best_card) {
                console.log('No best card found in data');
                container.innerHTML = `
                    <p>No cards found. <a href="/cards">Add your first card</a> to get recommendations.</p>
                `;
                return;
            }
            
            // Create container HTML based on whether to show all cards or just the best card
            let html = '';
            
            if (showAllCards) {
                html = `
                    <h3>Card Recommendations</h3>
                    <div class="results-container">
                        <div class="best-card">
                            <h4>Best Card to Use</h4>
                            <div id="best-card-result" class="card-result">
                                ${this.createCardResultHTML(data.best_card, amount, true)}
                            </div>
                        </div>
                        
                        <div class="other-cards">
                            <h4>Other Cards</h4>
                            <div id="other-cards-results" class="cards-results">
                `;
                
                // Add other cards
                if (data.all_cards && data.all_cards.length > 1) {
                    html += '<div class="other-cards-accordion">';
                    for (let i = 1; i < data.all_cards.length; i++) {
                        html += this.createCardResultHTML(data.all_cards[i], amount, false);
                    }
                    html += '</div>';
                } else {
                    html += '<p>No other cards available.</p>';
                }
                
                html += `
                            </div>
                        </div>
                    </div>
                `;
            } else {
                // Just show the best card
                html = `
                    <h3>Best Card to Use:</h3>
                    <div class="card-recommendation">
                        ${this.createCardResultHTML(data.best_card, amount, true)}
                    </div>
                `;
            }
            
            // Set the HTML
            container.innerHTML = html;
            
            // Show the container
            container.classList.remove('hidden');
            container.style.display = 'block';
            
            // Add event listeners for accordion if showing all cards
            if (showAllCards) {
                setTimeout(() => {
                    document.querySelectorAll('.card-result-header').forEach(header => {
                        if (!header.closest('.best-card')) {
                            header.addEventListener('click', function() {
                                console.log('Card header clicked for expansion');
                                const card = this.closest('.card-result-item');
                                card.classList.toggle('expanded');
                            });
                        }
                    });
                }, 0);
            }
        } catch (error) {
            console.error('Error in displayRecommendation:', error);
            container.innerHTML = `
                <div class="error-message">
                    <p>Error displaying recommendation results.</p>
                    <p>Error details: ${error.message}</p>
                </div>
            `;
        }
    },
    
    // Initialize recommendation form
    initializeRecommendationForm: function(formId, resultContainerId, showAllCards = true) {
        const form = document.getElementById(formId);
        const resultContainer = document.getElementById(resultContainerId);
        
        if (!form || !resultContainer) {
            console.error('Missing required elements:', { form, resultContainer });
            return;
        }
        
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            
            const merchant = form.querySelector('[name="merchant"]').value.toLowerCase();
            const category = form.querySelector('[name="category"]').value.toLowerCase();
            const amount = parseFloat(form.querySelector('[name="amount"]').value);
            
            if (!category) {
                Utils.showError('Please select a category');
                return;
            }
            
            if (isNaN(amount) || amount <= 0) {
                Utils.showError('Please enter a valid amount');
                return;
            }
            
            // Show loading state
            resultContainer.innerHTML = '<div class="loading">Calculating best cards...</div>';
            resultContainer.classList.remove('hidden');
            resultContainer.style.display = 'block';
            
            // Get user's cards
            const userCards = Storage.getCards().map(card => card.id);
            
            // Get recommendation from API
            this.getRecommendationFromAPI(
                merchant, 
                category, 
                amount, 
                userCards,
                (data, currentRecommendation) => {
                    // Success callback
                    this.displayRecommendation(resultContainer, data, amount, showAllCards);
                    
                    // Store current recommendation
                    form.dataset.currentRecommendation = JSON.stringify(currentRecommendation);
                },
                (error) => {
                    // Error callback
                    resultContainer.innerHTML = `
                        <div class="error-message">
                            <p>Sorry, we couldn't get recommendations right now. Please try again.</p>
                            <p>Error details: ${error.message}</p>
                        </div>
                    `;
                }
            );
        });
    }
};

// Initialize the recommendation page functionality
document.addEventListener('DOMContentLoaded', function() {
    // DOM Elements
    const recommendForm = document.getElementById('recommend-form');
    const recommendationResults = document.getElementById('recommendation-results');
    const saveTransactionBtn = document.getElementById('save-transaction-btn');
    
    if (recommendForm && recommendationResults) {
        // Initialize the full recommendation page
        RecommendationUI.initializeRecommendationForm('recommend-form', 'recommendation-results', true);
        
        // Transaction modal functionality
        const transactionModal = document.getElementById('transaction-modal');
        const saveTransactionForm = document.getElementById('save-transaction-form');
        const closeTransactionModal = document.querySelector('.close-transaction-modal');
        const cancelTransactionBtn = document.getElementById('cancel-transaction');
        const transactionCardSelect = document.getElementById('transaction-card');
        
        if (saveTransactionBtn && transactionModal && saveTransactionForm) {
            // Set today's date as default
            if (document.getElementById('transaction-date')) {
                document.getElementById('transaction-date').value = Utils.getTodayDate();
            }
            
            // Event listeners for transaction modal
            saveTransactionBtn.addEventListener('click', showSaveTransactionForm);
            closeTransactionModal.addEventListener('click', hideTransactionForm);
            cancelTransactionBtn.addEventListener('click', hideTransactionForm);
            saveTransactionForm.addEventListener('submit', saveTransaction);
            
            function showSaveTransactionForm() {
                const currentRecommendationData = recommendForm.dataset.currentRecommendation;
                if (!currentRecommendationData) {
                    return;
                }
                
                const currentRecommendation = JSON.parse(currentRecommendationData);
                
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
                const currentRecommendationData = recommendForm.dataset.currentRecommendation;
                if (!currentRecommendationData) return;
                
                const currentRecommendation = JSON.parse(currentRecommendationData);
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
        }
    }
});