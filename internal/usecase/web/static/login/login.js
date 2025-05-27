document.getElementById('login-btn').addEventListener('click', async function(event) {
    event.preventDefault();
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const errorLabel = document.getElementById('login-error');
    errorLabel.textContent = '';
    const response = await fetch('/api/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ username, password })
    });
    if (response.ok) {
        try {
            const data = await response.json();
            if (data && data.redirect) {
                window.location.href = data.redirect;
                return;
            }
        } catch (e) {}
        window.location.href = '/';
    } else {
        let errorMsg = 'Login failed';
        try {
            const data = await response.json();
            if (data && data.error) {
                errorMsg = data.error;
            }
        } catch (e) {}
        errorLabel.textContent = errorMsg;
    }
}); 