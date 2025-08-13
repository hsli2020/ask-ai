// Tab switching functionality
function showTab(tabName) {
    // Hide all tab contents
    const passwordTab = document.getElementById('password-login');
    const smsTab = document.getElementById('sms-login');
    
    passwordTab.classList.add('hidden');
    smsTab.classList.add('hidden');
    
    // Show selected tab
    document.getElementById(tabName).classList.remove('hidden');
    
    // Update tab button styles
    const passwordTabBtn = document.getElementById('password-tab');
    const smsTabBtn = document.getElementById('sms-tab');
    
    // Reset both buttons to inactive state
    passwordTabBtn.className = 'flex-1 py-2 px-4 text-sm font-medium rounded-md transition-colors duration-200 text-gray-500 hover:text-gray-700';
    smsTabBtn.className = 'flex-1 py-2 px-4 text-sm font-medium rounded-md transition-colors duration-200 text-gray-500 hover:text-gray-700';
    
    // Set active button
    if (tabName === 'password-login') {
        passwordTabBtn.className = 'flex-1 py-2 px-4 text-sm font-medium rounded-md transition-colors duration-200 bg-white text-primary shadow-sm';
    } else if (tabName === 'sms-login') {
        smsTabBtn.className = 'flex-1 py-2 px-4 text-sm font-medium rounded-md transition-colors duration-200 bg-white text-primary shadow-sm';
    }
}

// Toggle password visibility
function togglePassword(inputId) {
    const passwordInput = document.getElementById(inputId);
    const eyeIcon = document.getElementById(inputId + '-eye');
    
    if (passwordInput.type === 'password') {
        passwordInput.type = 'text';
        eyeIcon.innerHTML = `
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21" />
        `;
    } else {
        passwordInput.type = 'password';
        eyeIcon.innerHTML = `
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
        `;
    }
}

// Handle password login form submission
function handlePasswordLogin(event) {
    event.preventDefault();
    
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const rememberMe = document.getElementById('remember-me').checked;
    
    // Show loading state
    const submitBtn = event.target.querySelector('button[type="submit"]');
    const originalText = submitBtn.textContent;
    submitBtn.textContent = 'Signing In...';
    submitBtn.disabled = true;
    
    // Simulate API call
    setTimeout(() => {
        console.log('Password Login:', {
            username,
            password: '***hidden***',
            rememberMe
        });
        
        // Show success message (in real app, you'd redirect or update UI)
        alert(`Welcome back, ${username}!`);
        
        // Reset button state
        submitBtn.textContent = originalText;
        submitBtn.disabled = false;
    }, 1500);
}

// Handle SMS login form submission
function handleSmsLogin(event) {
    event.preventDefault();
    
    const phone = document.getElementById('phone').value;
    const smsCode = document.getElementById('sms-code').value;
    
    // Show loading state
    const submitBtn = event.target.querySelector('button[type="submit"]');
    const originalText = submitBtn.textContent;
    submitBtn.textContent = 'Verifying...';
    submitBtn.disabled = true;
    
    // Simulate API call
    setTimeout(() => {
        console.log('SMS Login:', {
            phone,
            smsCode
        });
        
        // Show success message (in real app, you'd redirect or update UI)
        alert(`Successfully logged in with phone number: ${phone}`);
        
        // Reset button state
        submitBtn.textContent = originalText;
        submitBtn.disabled = false;
    }, 1500);
}

// Send SMS code functionality
let smsCodeSent = false;
let countdown = 0;

function sendSmsCode() {
    const phoneInput = document.getElementById('phone');
    const sendCodeBtn = document.getElementById('send-code-btn');
    
    if (!phoneInput.value) {
        alert('Please enter a phone number first');
        phoneInput.focus();
        return;
    }
    
    if (countdown > 0) {
        return; // Still in cooldown
    }
    
    // Show loading state
    sendCodeBtn.textContent = 'Sending...';
    sendCodeBtn.disabled = true;
    
    // Simulate sending SMS
    setTimeout(() => {
        console.log('SMS code sent to:', phoneInput.value);
        alert('SMS code sent! Check your phone.');
        
        smsCodeSent = true;
        countdown = 60; // 60 second cooldown
        
        // Start countdown
        const countdownInterval = setInterval(() => {
            sendCodeBtn.textContent = `Resend (${countdown}s)`;
            countdown--;
            
            if (countdown < 0) {
                clearInterval(countdownInterval);
                sendCodeBtn.textContent = 'Resend Code';
                sendCodeBtn.disabled = false;
                countdown = 0;
            }
        }, 1000);
        
    }, 1000);
}

// Add input validation and formatting
document.addEventListener('DOMContentLoaded', function() {
    // Format phone number input
    const phoneInput = document.getElementById('phone');
    if (phoneInput) {
        phoneInput.addEventListener('input', function(e) {
            let value = e.target.value.replace(/\D/g, '');
            if (value.length >= 10) {
                value = value.substring(0, 10);
                // Format as (XXX) XXX-XXXX
                value = `(${value.substring(0, 3)}) ${value.substring(3, 6)}-${value.substring(6)}`;
            }
            e.target.value = value;
        });
    }
    
    // Format SMS code input (only numbers, max 6 digits)
    const smsCodeInput = document.getElementById('sms-code');
    if (smsCodeInput) {
        smsCodeInput.addEventListener('input', function(e) {
            let value = e.target.value.replace(/\D/g, '');
            if (value.length > 6) {
                value = value.substring(0, 6);
            }
            e.target.value = value;
        });
    }
    
    // The focus ring is already handled by Tailwind CSS classes in the HTML
    // No need for additional JavaScript focus handling
});

// Handle social login buttons
document.addEventListener('DOMContentLoaded', function() {
    const socialButtons = document.querySelectorAll('button');
    socialButtons.forEach(button => {
        if (button.textContent.includes('Google') || button.textContent.includes('Facebook')) {
            button.addEventListener('click', function() {
                const provider = this.textContent.trim();
                alert(`${provider} login would be implemented here`);
                console.log(`${provider} login clicked`);
            });
        }
    });
});