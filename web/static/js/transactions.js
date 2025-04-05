// Transactions management functionality

document.addEventListener('DOMContentLoaded', function() {
    // DOM Elements
    const transactionsList = document.getElementById('transactions-list');
    const addTransactionBtn = document.getElementById('add-transaction-btn');
    const transactionFormModal = document.getElementById('transaction-form-modal');
    const transactionForm = document.getElementById('transaction-form');
    const cancelFormBtn = document.getElementById('cancel-form');
    const closeModalBtn = document.querySelector('.close-modal');
    const formTitle = document.getElementById('transaction-form-title');
    
    const filterDateFrom = document.getElementById('filter-date-from');
    const filterDateTo = document.getElementById('filter-date-to');
    const filterCategory = document.getElementById('filter-category');
    const filterCard = document.getElementById('filter-card');
    const applyFiltersBtn = document.getElementById('apply-filters');
    
    const totalSpentElement = document.getElementById('total-spent');
    const totalRewardsElement = document.getElementById('total-rewards');
    const avgRewardRateElement = document.getElementById('avg-reward-rate');
    
    // Current filters
    let currentFilters = {
        dateFrom: '',
        dateTo: '',
        category: '',
        cardId: ''
    };
    
    // Initialize
    loadCards();
    loadTransactions();
    
    // Set today's date as default
    const today = Utils.getTodayDate();
    document.getElementById('transaction-date').value = today;
    
    // Event Listeners
    addTransactionBtn.addEventListener('click', showAddTransactionForm);
    cancelFormBtn.addEventListener('click', hideTransactionForm);
    closeModalBtn.addEventListener('click', hideTransactionForm);
    transactionForm.addEventListener('submit', saveTransaction);
    
    applyFiltersBtn.addEventListener('click', applyFilters);
    
    // Functions
    function loadCards() {
        const cards = Storage.getCards();
        let cardOptions = '<option value="">All Cards</option>';
        
        cards.forEach(card => {
            cardOptions += `<option value="${card.id}">${card.name} (*${card.last4Digits})</option>`;
        });
        
        // Populate filter card select
        filterCard.innerHTML = cardOptions;
        
        // Populate transaction form card select
        const transactionCardSelect = document.getElementById('transaction-card');
        if (transactionCardSelect) {
            transactionCardSelect.innerHTML = cardOptions.replace('All Cards', 'Select a Card');
        }
    }
    
    function loadTransactions() {
        const transactions = Storage.getFilteredTransactions(currentFilters);
        
        if (transactions.length === 0) {
            transactionsList.innerHTML = '<p>No transactions recorded yet.</p>';
            updateSummary(transactions);
            return;
        }
        
        // Sort transactions by date (newest first)
        transactions.sort((a, b) => new Date(b.date) - new Date(a.date));
        
        let html = '';
        transactions.forEach(transaction => {
            const card = Storage.getCardById(transaction.cardId);
            const cardName = card ? `${card.name} (*${card.last4Digits})` : 'Unknown Card';
            
            html += `
                <div class="transaction-item" data-id="${transaction.id}">
                    <div class="transaction-info">
                        <div class="transaction-merchant">${transaction.merchantName}</div>
                        <div class="transaction-details">
                            <span>${Utils.formatDate(transaction.date)}</span> • 
                            <span>${transaction.category}</span> • 
                            <span>${cardName}</span>
                            ${transaction.notes ? `<div class="transaction-notes">${transaction.notes}</div>` : ''}
                        </div>
                    </div>
                    <div class="transaction-values">
                        <div class="transaction-amount">${Utils.formatCurrency(transaction.amount)}</div>
                        <div class="transaction-reward">+${Utils.formatCurrency(transaction.rewardEarned)}</div>
                    </div>
                    <div class="transaction-actions">
                        <button class="transaction-action edit-transaction" title="Edit Transaction">
                            <i class="fas fa-edit">✏️</i>
                        </button>
                        <button class="transaction-action delete-transaction" title="Delete Transaction">
                            <i class="fas fa-trash">🗑️</i>
                        </button>
                    </div>
                </div>
            `;
        });
        
        transactionsList.innerHTML = html;
        
        // Update summary
        updateSummary(transactions);
        
        // Add event listeners to transaction actions
        document.querySelectorAll('.edit-transaction').forEach(btn => {
            btn.addEventListener('click', function() {
                const transactionId = parseInt(this.closest('.transaction-item').dataset.id);
                showEditTransactionForm(transactionId);
            });
        });
        
        document.querySelectorAll('.delete-transaction').forEach(btn => {
            btn.addEventListener('click', function() {
                const transactionId = parseInt(this.closest('.transaction-item').dataset.id);
                deleteTransaction(transactionId);
            });
        });
    }
    
    function updateSummary(transactions) {
        if (transactions.length === 0) {
            totalSpentElement.textContent = Utils.formatCurrency(0);
            totalRewardsElement.textContent = Utils.formatCurrency(0);
            avgRewardRateElement.textContent = '0.00%';
            return;
        }
        
        const totalSpent = transactions.reduce((sum, t) => sum + t.amount, 0);
        const totalRewards = transactions.reduce((sum, t) => sum + t.rewardEarned, 0);
        const avgRewardRate = (totalRewards / totalSpent) * 100;
        
        totalSpentElement.textContent = Utils.formatCurrency(totalSpent);
        totalRewardsElement.textContent = Utils.formatCurrency(totalRewards);
        avgRewardRateElement.textContent = avgRewardRate.toFixed(2) + '%';
    }
    
    function showAddTransactionForm() {
        formTitle.textContent = 'Add Transaction';
        transactionForm.reset();
        transactionForm.elements['id'].value = '';
        transactionForm.elements['date'].value = Utils.getTodayDate();
        
        Utils.toggleModal('transaction-form-modal', true);
    }
    
    function showEditTransactionForm(transactionId) {
        const transaction = Storage.getTransactionById(transactionId);
        if (!transaction) return;
        
        formTitle.textContent = 'Edit Transaction';
        
        const form = transactionForm;
        form.elements['id'].value = transaction.id;
        form.elements['date'].value = transaction.date;
        form.elements['merchantName'].value = transaction.merchantName;
        form.elements['category'].value = transaction.category;
        form.elements['amount'].value = transaction.amount;
        form.elements['cardId'].value = transaction.cardId;
        form.elements['rewardEarned'].value = transaction.rewardEarned;
        form.elements['notes'].value = transaction.notes || '';
        
        Utils.toggleModal('transaction-form-modal', true);
    }
    
    function hideTransactionForm() {
        Utils.toggleModal('transaction-form-modal', false);
    }
    
    function saveTransaction(e) {
        e.preventDefault();
        
        const form = transactionForm;
        const id = form.elements['id'].value ? parseInt(form.elements['id'].value) : null;
        
        const transaction = {
            id: id,
            date: form.elements['date'].value,
            merchantName: form.elements['merchantName'].value,
            category: form.elements['category'].value,
            amount: parseFloat(form.elements['amount'].value),
            cardId: parseInt(form.elements['cardId'].value),
            rewardEarned: parseFloat(form.elements['rewardEarned'].value),
            notes: form.elements['notes'].value
        };
        
        if (id) {
            Storage.updateTransaction(transaction);
        } else {
            Storage.addTransaction(transaction);
        }
        
        hideTransactionForm();
        loadTransactions();
    }
    
    function deleteTransaction(transactionId) {
        if (confirm('Are you sure you want to delete this transaction?')) {
            Storage.deleteTransaction(transactionId);
            loadTransactions();
        }
    }
    
    function applyFilters() {
        currentFilters = {
            dateFrom: filterDateFrom.value,
            dateTo: filterDateTo.value,
            category: filterCategory.value,
            cardId: filterCard.value
        };
        
        loadTransactions();
    }
});
