const API_URL = "http://localhost:8099/notify";

const createForm = document.getElementById('createForm');
const getStatusForm = document.getElementById('getStatusForm');
const deleteForm = document.getElementById('deleteForm');
const responseOutput = document.getElementById('responseOutput');

async function sendRequest(url, method, body = null) {
    const options = {
        method: method,
        headers: {
            'Content-Type': 'application/json'
        },
    };
    if (body) {
        options.body = JSON.stringify(body);
    }

    try {
        const response = await fetch(url, options);
        const data = await response.json();
        responseOutput.textContent = JSON.stringify(data, null, 2);
    } catch (error) {
        responseOutput.textContent = `Ошибка: ${error.message}`;
    }
}

createForm.addEventListener('submit', (e) => {
    e.preventDefault();
    const recipientID = parseInt(document.getElementById('recipientID').value, 10);
    const date = document.getElementById('date').value;
    const text = document.getElementById('text').value;

    const body = { recipient_id: recipientID, date, text };
    sendRequest(API_URL, 'POST', body);
});

getStatusForm.addEventListener('submit', (e) => {
    e.preventDefault();
    const notificationID = document.getElementById('statusID').value;
    sendRequest(`${API_URL}/${notificationID}`, 'GET');
});

deleteForm.addEventListener('submit', (e) => {
    e.preventDefault();
    const notificationID = document.getElementById('deleteID').value;
    sendRequest(`${API_URL}/${notificationID}`, 'DELETE');
});