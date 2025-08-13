function showTab(tabName) {
    const tabContents = document.querySelectorAll('.tab-content');
    tabContents.forEach(tabContent => {
        tabContent.classList.remove('active');
    });

    const tabButtons = document.querySelectorAll('.tab-button');
    tabButtons.forEach(tabButton => {
        tabButton.classList.remove('active');
    });

    document.getElementById(tabName).classList.add('active');
    document.querySelector(`[onclick="showTab('${tabName}')"]`).classList.add('active');
}
