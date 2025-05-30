// Cards management functionality

document.addEventListener('DOMContentLoaded', function() {
    // DOM Elements
    const cardsList = document.getElementById('cards-list');
    const addCardBtn = document.getElementById('add-card-btn');
    const addPredefinedCardBtn = document.getElementById('add-predefined-card-btn');
    const cardFormModal = document.getElementById('card-form-modal');
    const cardForm = document.getElementById('card-form');
    const cancelFormBtn = document.getElementById('cancel-form');
    const closeModalBtn = document.querySelector('.close-modal');
    const formTitle = document.getElementById('form-title');
    
    const predefinedCardsSection = document.getElementById('predefined-cards-section');
    const predefinedCardsList = document.getElementById('predefined-cards-list');
    const closePredefinedCardsBtn = document.getElementById('close-predefined-cards');
    
    const rewardRulesSection = document.getElementById('reward-rules-section');
    const selectedCardName = document.getElementById('selected-card-name');
    const rulesList = document.getElementById('rules-list');
    const addRuleBtn = document.getElementById('add-rule-btn');
    const ruleFormModal = document.getElementById('rule-form-modal');
    const ruleForm = document.getElementById('rule-form');
    const cancelRuleFormBtn = document.getElementById('cancel-rule-form');
    const closeRuleModalBtn = document.querySelector('.close-rule-modal');
    const ruleFormTitle = document.getElementById('rule-form-title');
    const rewardTypeSelect = document.getElementById('reward-type');
    const pointValueGroup = document.getElementById('point-value-group');
    
    // Current card ID for reward rules
    let currentCardId = null;
    
    // Initialize
    loadCards();
    
    // Event Listeners
    addCardBtn.addEventListener('click', showAddCardForm);
    addPredefinedCardBtn.addEventListener('click', showPredefinedCards);
    cancelFormBtn.addEventListener('click', hideCardForm);
    closeModalBtn.addEventListener('click', hideCardForm);
    cardForm.addEventListener('submit', saveCard);
    closePredefinedCardsBtn.addEventListener('click', hidePredefinedCards);
    
    addRuleBtn.addEventListener('click', showAddRuleForm);
    cancelRuleFormBtn.addEventListener('click', hideRuleForm);
    closeRuleModalBtn.addEventListener('click', hideRuleForm);
    ruleForm.addEventListener('submit', saveRule);
    
    rewardTypeSelect.addEventListener('change', togglePointValueVisibility);
    
    // Functions
    
    // Load predefined cards from the server
    function loadPredefinedCards() {
        fetch('/api/cards')
            .then(response => response.json())
            .then(cards => {
                if (cards.length === 0) {
                    predefinedCardsList.innerHTML = '<p>No predefined cards available.</p>';
                    return;
                }
                
                let html = '';
                cards.forEach(card => {
                    // Create reward rules HTML
                    let rulesHtml = '';
                    if (card.reward_rules && card.reward_rules.length > 0) {
                        rulesHtml += '<div class="card-reward-rules">';
                        rulesHtml += '<h4>Reward Rules:</h4>';
                        rulesHtml += '<ul>';
                        card.reward_rules.forEach(rule => {
                            rulesHtml += `<li>${rule.type}: ${rule.entity_name} - ${rule.reward_rate}% ${rule.reward_type}</li>`;
                        });
                        rulesHtml += '</ul>';
                        rulesHtml += '</div>';
                    }
                    
                    // Create benefits HTML
                    let benefitsHtml = '';
                    if (card.benefits && card.benefits.length > 0) {
                        benefitsHtml += '<div class="card-benefits">';
                        benefitsHtml += '<h4>Benefits:</h4>';
                        benefitsHtml += '<ul>';
                        card.benefits.forEach(benefit => {
                            benefitsHtml += `<li>${benefit}</li>`;
                        });
                        benefitsHtml += '</ul>';
                        benefitsHtml += '</div>';
                    }
                    
                    html += `
                        <div class="predefined-card-item" data-key="${card.card_key}">
                            <div class="card-header">
                                <div class="card-title">${card.name}</div>
                                <button class="btn btn-primary add-this-card" data-key="${card.card_key}">Add This Card</button>
                            </div>
                            <div class="card-details">
                                <div class="card-detail">Issuer: ${card.issuer}</div>
                                <div class="card-detail">Type: ${card.card_type}</div>
                                <div class="card-detail">Default Reward: ${card.default_reward_rate}% ${card.reward_type}</div>
                                <div class="card-detail">Point Value: ₹${card.point_value}</div>
                                <div class="card-detail">Annual Fee: ₹${card.annual_fee}</div>
                                <div class="card-detail">Fee Waiver: ${card.annual_fee_waiver || 'None'}</div>
                            </div>
                            ${rulesHtml}
                            ${benefitsHtml}
                        </div>
                    `;
                });
                
                predefinedCardsList.innerHTML = html;
                
                // Add event listeners to "Add This Card" buttons
                document.querySelectorAll('.add-this-card').forEach(btn => {
                    btn.addEventListener('click', function() {
                        const cardKey = this.dataset.key;
                        showAddPredefinedCardForm(cardKey);
                    });
                });
            })
            .catch(error => {
                console.error('Error loading predefined cards:', error);
                predefinedCardsList.innerHTML = '<p>Error loading predefined cards. Please try again later.</p>';
            });
    }
    
    function showPredefinedCards() {
        predefinedCardsSection.classList.remove('hidden');
        loadPredefinedCards();
    }
    
    function hidePredefinedCards() {
        predefinedCardsSection.classList.add('hidden');
    }
    
    function showAddPredefinedCardForm(cardKey) {
        fetch(`/api/cards/${cardKey}`)
            .then(response => response.json())
            .then(card => {
                formTitle.textContent = `Add ${card.name}`;
                
                const form = cardForm;
                form.reset();
                form.elements['id'].value = '';
                form.elements['name'].value = card.name;
                form.elements['issuer'].value = card.issuer;
                form.elements['cardType'].value = card.card_type;
                form.elements['defaultRewardRate'].value = card.default_reward_rate;
                
                // Hide predefined cards section
                hidePredefinedCards();
                
                // Show the form modal
                Utils.toggleModal('card-form-modal', true);
                
                // Store the predefined card key for later use
                form.dataset.predefinedCardKey = cardKey;
            })
            .catch(error => {
                console.error('Error loading predefined card details:', error);
                Utils.showError('Error loading predefined card details. Please try again.');
            });
    }
    
    function loadCards() {
        const cards = Storage.getCards();
        
        if (cards.length === 0) {
            cardsList.innerHTML = '<p>No cards added yet. Add your first card to get started.</p>';
            return;
        }
        
        let html = '';
        cards.forEach(card => {
            html += `
                <div class="card-item" data-id="${card.id}">
                    <div class="card-header">
                        <div class="card-title">${card.name}</div>
                        <div class="card-actions">
                            <button class="card-action edit-card" title="Edit Card">
                                <i class="fas fa-edit">✏️</i>
                            </button>
                            <button class="card-action delete-card" title="Delete Card">
                                <i class="fas fa-trash">🗑️</i>
                            </button>
                            <button class="card-action view-rules" title="View Reward Rules">
                                <i class="fas fa-list">📋</i>
                            </button>
                        </div>
                    </div>
                    <div class="card-details">
                        <div class="card-detail">Issuer: ${card.issuer}</div>
                        <div class="card-detail">Card Number: **** **** **** ${card.last4Digits}</div>
                        <div class="card-detail">Expires: ${formatExpiryDate(card.expiryDate)}</div>
                        <div class="card-detail">Type: ${card.cardType}</div>
                        <div class="card-detail">Default Reward: ${card.defaultRewardRate}%</div>
                    </div>
                </div>
            `;
        });
        
        cardsList.innerHTML = html;
        
        // Add event listeners to card actions
        document.querySelectorAll('.edit-card').forEach(btn => {
            btn.addEventListener('click', function() {
                const cardId = parseInt(this.closest('.card-item').dataset.id);
                showEditCardForm(cardId);
            });
        });
        
        document.querySelectorAll('.delete-card').forEach(btn => {
            btn.addEventListener('click', function() {
                const cardId = parseInt(this.closest('.card-item').dataset.id);
                deleteCard(cardId);
            });
        });
        
        document.querySelectorAll('.view-rules').forEach(btn => {
            btn.addEventListener('click', function() {
                const cardId = parseInt(this.closest('.card-item').dataset.id);
                showCardRules(cardId);
            });
        });
    }
    
    function formatExpiryDate(dateString) {
        const date = new Date(dateString);
        return `${date.getMonth() + 1}/${date.getFullYear().toString().substr(-2)}`;
    }
    
    function showAddCardForm() {
        formTitle.textContent = 'Add New Card';
        cardForm.reset();
        cardForm.elements['id'].value = '';
        Utils.toggleModal('card-form-modal', true);
    }
    
    function showEditCardForm(cardId) {
        const card = Storage.getCardById(cardId);
        if (!card) return;
        
        formTitle.textContent = 'Edit Card';
        
        const form = cardForm;
        form.elements['id'].value = card.id;
        form.elements['name'].value = card.name;
        form.elements['issuer'].value = card.issuer;
        form.elements['last4Digits'].value = card.last4Digits;
        
        // Format date for input[type=month]
        const expiryDate = new Date(card.expiryDate);
        const year = expiryDate.getFullYear();
        const month = String(expiryDate.getMonth() + 1).padStart(2, '0');
        form.elements['expiryDate'].value = `${year}-${month}`;
        
        form.elements['cardType'].value = card.cardType;
        form.elements['defaultRewardRate'].value = card.defaultRewardRate;
        
        Utils.toggleModal('card-form-modal', true);
    }
    
    function hideCardForm() {
        Utils.toggleModal('card-form-modal', false);
    }
    
    function saveCard(e) {
        e.preventDefault();
        
        const form = cardForm;
        const id = form.elements['id'].value ? parseInt(form.elements['id'].value) : null;
        
        // Parse expiry date
        const expiryInput = form.elements['expiryDate'].value;
        const [year, month] = expiryInput.split('-');
        const expiryDate = new Date(parseInt(year), parseInt(month) - 1, 1);
        
        const card = {
            id: id,
            name: form.elements['name'].value,
            issuer: form.elements['issuer'].value,
            last4Digits: form.elements['last4Digits'].value,
            expiryDate: expiryDate,
            cardType: form.elements['cardType'].value,
            defaultRewardRate: parseFloat(form.elements['defaultRewardRate'].value),
            rewardType: 'Cashback',  // Default to cashback, can be changed with reward rules
            pointValue: 1            // Default point value
        };
        
        if (id) {
            Storage.updateCard(card);
        } else {
            const newCard = Storage.addCard(card);
            
            // If this was a predefined card, also add its reward rules
            const predefinedCardKey = form.dataset.predefinedCardKey;
            if (predefinedCardKey) {
                fetch(`/api/cards/${predefinedCardKey}`)
                    .then(response => response.json())
                    .then(predefinedCard => {
                        if (predefinedCard.reward_rules && predefinedCard.reward_rules.length > 0) {
                            predefinedCard.reward_rules.forEach(rule => {
                                const newRule = {
                                    cardId: newCard.id,
                                    type: rule.type,
                                    entityName: rule.entity_name,
                                    rewardRate: rule.reward_rate,
                                    rewardType: rule.reward_type,
                                    pointValue: predefinedCard.point_value
                                };
                                Storage.addRewardRule(newRule);
                            });
                        }
                    })
                    .catch(error => {
                        console.error('Error loading predefined card rules:', error);
                    });
                
                // Clear the predefined card key
                delete form.dataset.predefinedCardKey;
            }
        }
        
        hideCardForm();
        loadCards();
    }
    
    function deleteCard(cardId) {
        if (confirm('Are you sure you want to delete this card? This will also delete all associated reward rules.')) {
            Storage.deleteCard(cardId);
            loadCards();
            
            // Hide reward rules section if it's showing the deleted card
            if (currentCardId === cardId) {
                rewardRulesSection.classList.add('hidden');
                currentCardId = null;
            }
        }
    }
    
    // Reward Rules Functions
    function showCardRules(cardId) {
        const card = Storage.getCardById(cardId);
        if (!card) return;
        
        currentCardId = cardId;
        selectedCardName.textContent = card.name;
        rewardRulesSection.classList.remove('hidden');
        
        loadCardRules(cardId);
    }
    
    function loadCardRules(cardId) {
        const rules = Storage.getCardRules(cardId);
        
        if (rules.length === 0) {
            rulesList.innerHTML = '<p>No reward rules added for this card yet.</p>';
            return;
        }
        
        let html = '';
        rules.forEach(rule => {
            let rewardDisplay = `${rule.rewardRate}% ${rule.rewardType}`;
            if (rule.rewardType === 'Points' || rule.rewardType === 'Miles') {
                rewardDisplay += ` (₹${rule.pointValue} per point)`;
            }
            
            html += `
                <div class="rule-item" data-id="${rule.id}">
                    <div class="rule-header">
                        <div class="rule-title">${rule.type}: ${rule.entityName}</div>
                        <div class="rule-actions">
                            <button class="rule-action edit-rule" title="Edit Rule">
                                <i class="fas fa-edit">✏️</i>
                            </button>
                            <button class="rule-action delete-rule" title="Delete Rule">
                                <i class="fas fa-trash">🗑️</i>
                            </button>
                        </div>
                    </div>
                    <div class="rule-details">
                        <div class="rule-detail">Reward: ${rewardDisplay}</div>
                    </div>
                </div>
            `;
        });
        
        rulesList.innerHTML = html;
        
        // Add event listeners to rule actions
        document.querySelectorAll('.edit-rule').forEach(btn => {
            btn.addEventListener('click', function() {
                const ruleId = parseInt(this.closest('.rule-item').dataset.id);
                showEditRuleForm(ruleId);
            });
        });
        
        document.querySelectorAll('.delete-rule').forEach(btn => {
            btn.addEventListener('click', function() {
                const ruleId = parseInt(this.closest('.rule-item').dataset.id);
                deleteRule(ruleId);
            });
        });
    }
    
    function showAddRuleForm() {
        ruleFormTitle.textContent = 'Add Reward Rule';
        ruleForm.reset();
        ruleForm.elements['id'].value = '';
        ruleForm.elements['cardId'].value = currentCardId;
        
        // Default point value
        ruleForm.elements['pointValue'].value = '0.25';
        
        togglePointValueVisibility();
        Utils.toggleModal('rule-form-modal', true);
    }
    
    function showEditRuleForm(ruleId) {
        const rules = Storage.getRewardRules();
        const rule = rules.find(r => r.id === ruleId);
        if (!rule) return;
        
        ruleFormTitle.textContent = 'Edit Reward Rule';
        
        const form = ruleForm;
        form.elements['id'].value = rule.id;
        form.elements['cardId'].value = rule.cardId;
        form.elements['type'].value = rule.type;
        form.elements['entityName'].value = rule.entityName;
        form.elements['rewardRate'].value = rule.rewardRate;
        form.elements['rewardType'].value = rule.rewardType;
        form.elements['pointValue'].value = rule.pointValue || '0.25';
        
        togglePointValueVisibility();
        Utils.toggleModal('rule-form-modal', true);
    }
    
    function hideRuleForm() {
        Utils.toggleModal('rule-form-modal', false);
    }
    
    function togglePointValueVisibility() {
        const rewardType = rewardTypeSelect.value;
        if (rewardType === 'Points' || rewardType === 'Miles') {
            pointValueGroup.style.display = 'block';
        } else {
            pointValueGroup.style.display = 'none';
        }
    }
    
    function saveRule(e) {
        e.preventDefault();
        
        const form = ruleForm;
        const id = form.elements['id'].value ? parseInt(form.elements['id'].value) : null;
        const cardId = parseInt(form.elements['cardId'].value);
        const rewardType = form.elements['rewardType'].value;
        
        let pointValue = null;
        if (rewardType === 'Points' || rewardType === 'Miles') {
            pointValue = parseFloat(form.elements['pointValue'].value);
        }
        
        const rule = {
            id: id,
            cardId: cardId,
            type: form.elements['type'].value,
            entityName: form.elements['entityName'].value,
            rewardRate: parseFloat(form.elements['rewardRate'].value),
            rewardType: rewardType,
            pointValue: pointValue
        };
        
        if (id) {
            Storage.updateRewardRule(rule);
        } else {
            Storage.addRewardRule(rule);
        }
        
        hideRuleForm();
        loadCardRules(cardId);
    }
    
    function deleteRule(ruleId) {
        if (confirm('Are you sure you want to delete this reward rule?')) {
            Storage.deleteRewardRule(ruleId);
            loadCardRules(currentCardId);
        }
    }
});
