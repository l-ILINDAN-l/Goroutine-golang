const orderInput = document.getElementById('orderIdInput');
const searchButton = document.getElementById('searchButton');
const resultOutput = document.getElementById('resultOutput');

orderInput.addEventListener('input', () => {
    if (orderInput.value.trim() !== '') {
        searchButton.disabled = false;
    } else {
        searchButton.disabled = true;
    }
});


searchButton.addEventListener('click', async () => {
    const orderId = orderInput.value;
    if (!orderId) {
        resultOutput.textContent = 'Пожалуйста, введите ID заказа.';
        return;
    }

    try {
        const response = await fetch(`/api/v1/order/${orderId}`);

        const data = await response.json();

        if (!response.ok) {
            resultOutput.textContent = `Ошибка: ${data.error || response.statusText}`;
        } else {
            resultOutput.textContent = JSON.stringify(data, null, 2);
        }
    } catch (error) {
        resultOutput.textContent = `Сетевая ошибка: ${error.message}`;
    }
});