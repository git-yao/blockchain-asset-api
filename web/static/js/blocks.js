class BlockExplorer {
    constructor() {
        this.currentPage = 1;
        this.pageSize = 10;
        this.init();
    }

    init() {
        document.addEventListener('DOMContentLoaded', () => {
            this.loadTransactions();
        });
    }

    // 加载交易数据
    async loadTransactions(page = 1) {
        this.currentPage = page;

        const params = new URLSearchParams({
            page: page,
            size: this.pageSize,
            tx_type: document.getElementById('txType').value,
            address: document.getElementById('address').value,
            block_number: document.getElementById('blockNumber').value
        });

        try {
            const response = await fetch(`/api/v1/transactions?${params}`);
            const data = await response.json();
            this.renderTransactions(data.transactions);
            this.renderPagination(data.total, data.page, data.pages);
        } catch (error) {
            console.error('Error loading transactions:', error);
            alert('加载数据失败');
        }
    }

    // 渲染交易数据
    renderTransactions(transactions) {
        const tbody = document.getElementById('transactionsTable');
        tbody.innerHTML = '';

        if (!transactions || transactions.length === 0) {
            tbody.innerHTML = '<tr><td colspan="9" class="text-center">暂无数据</td></tr>';
            return;
        }
        // console.log(transactions);
        transactions.forEach(tx => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td><a href="#" class="tx-hash" onclick="blockExplorer.showTransactionDetail('${tx.tx_hash}')">${this.formatHash(tx.tx_hash)}</a></td>
                <td>${tx.block_number}</td>
                <td><span class="address-hash" title="${tx.from_address}">${this.formatHash(tx.from_address)}</span></td>
                <td><span class="address-hash" title="${tx.to_address ? tx.to_address : '合约创建'}">${tx.to_address ? this.formatHash(tx.to_address) : '合约创建'}</span></td>
                <td class="amount-value">${tx.tx_type === 'erc20_transfer' ?  (tx.erc20_amount + 'Token') : (tx.value + 'ETH')}</td>
                <td class="gas-fee">${this.calculateGasFee(tx.gas_price, tx.gas_used)} ETH</td>
                <td><span class="transaction-type ${this.getTxTypeClass(tx.tx_type)}">${this.getTxTypeText(tx.tx_type)}</span></td>
                <td><span class="${tx.status === 'success' ? 'status-success' : 'status-failed'}">${tx.status === 'success' ? '成功' : '失败'}</span></td>
                <td>${tx.created_at ? this.formatDate(tx.created_at) : ''}</td>
            `;
            tbody.appendChild(row);
        });
    }

    // 渲染分页控件
    renderPagination(total, currentPage, totalPages) {
        const pagination = document.getElementById('pagination');
        pagination.innerHTML = '';

        if (totalPages <= 1) return;

        // 上一页
        const prevLi = document.createElement('li');
        prevLi.className = `page-item ${currentPage === 1 ? 'disabled' : ''}`;
        prevLi.innerHTML = `<a class="page-link" href="#" onclick="blockExplorer.loadTransactions(${Math.max(1, currentPage - 1)})">上一页</a>`;
        pagination.appendChild(prevLi);

        // 页码
        const startPage = Math.max(1, currentPage - 2);
        const endPage = Math.min(totalPages, currentPage + 2);

        for (let i = startPage; i <= endPage; i++) {
            const li = document.createElement('li');
            li.className = `page-item ${i === currentPage ? 'active' : ''}`;
            li.innerHTML = `<a class="page-link" href="#" onclick="blockExplorer.loadTransactions(${i})">${i}</a>`;
            pagination.appendChild(li);
        }

        // 下一页
        const nextLi = document.createElement('li');
        nextLi.className = `page-item ${currentPage === totalPages ? 'disabled' : ''}`;
        nextLi.innerHTML = `<a class="page-link" href="#" onclick="blockExplorer.loadTransactions(${Math.min(totalPages, currentPage + 1)})">下一页</a>`;
        pagination.appendChild(nextLi);
    }

    // 搜索功能
    searchTransactions() {
        this.loadTransactions(1);
    }

    // 重置筛选条件
    resetFilters() {
        document.getElementById('txType').value = '';
        document.getElementById('address').value = '';
        document.getElementById('blockNumber').value = '';
        this.loadTransactions(1);
    }

    // 显示交易详情
    showTransactionDetail(txHash) {
        alert(`交易详情: ${txHash}\n（此处可跳转到详细页面）`);
    }

    // 辅助函数
    formatHash(hash) {
        if (!hash) return '';
        return `${hash.substring(0, 6)}...${hash.substring(hash.length - 4)}`;
    }

    calculateGasFee(gasPrice, gasUsed) {
        if (!gasPrice || !gasUsed) return '0';
        // 简化处理，实际应该进行精确计算
        return (parseFloat(gasPrice) * parseInt(gasUsed)).toFixed(8);
    }

    getTxTypeClass(type) {
        switch(type) {
            case 'eth_transfer': return 'eth-transfer';
            case 'erc20_transfer': return 'erc20-transfer';
            case 'contract_call': return 'contract-call';
            default: return '';
        }
    }

    getTxTypeText(type) {
        switch(type) {
            case 'eth_transfer': return 'ETH转账';
            case 'erc20_transfer': return 'ERC20转账';
            case 'contract_call': return '合约调用';
            default: return type;
        }
    }

    formatDate(dateStr) {
        if (!dateStr) return '';
        const date = new Date(dateStr);
        return date.toLocaleString('zh-CN');
    }
}

// 创建全局实例
const blockExplorer = new BlockExplorer();

// 全局函数供HTML调用
function searchTransactions() {
    blockExplorer.searchTransactions();
}

function resetFilters() {
    blockExplorer.resetFilters();
}
