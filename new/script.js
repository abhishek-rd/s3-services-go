window.addEventListener("DOMContentLoaded", () => {
    fetch("/get-json")
        .then((response) => response.json())
        .then((data) => {
            populateTable(data.testSuites);
        });

    // Add event listener for the add-test-case button
    document.getElementById('add-test-case-button').addEventListener('click', openPopup);

    // Add event listener for the popup close button
    document.querySelector('.popup .close-button').addEventListener('click', closePopup);

    // Add event listeners for the options
    document.getElementById('option1').addEventListener('change', handleOptionSelection);
    document.getElementById('option2').addEventListener('change', handleOptionSelection);

    // Add event listener for the form submission
    document.getElementById('new-test-case-form').addEventListener('submit', createTestCase);
});

// Populate table function
function populateTable(testSuites) {
    // Add your logic to populate the table with test suites
}

function openPopup() {
    const popup = document.getElementById('popup');
    popup.style.display = 'block';
}

function closePopup() {
    const popup = document.getElementById('popup');
    popup.style.display = 'none';
}

function handleOptionSelection() {
    const option1Radio = document.getElementById('option1');
    const option2Radio = document.getElementById('option2');
    const uploadContainer = document.getElementById('upload-container');

    uploadContainer.innerHTML = '';

    if (option1Radio.checked) {
        uploadContainer.innerHTML = `
        <label for="csv-file">CSV File:</label>
        <input type="file" id="csv-file" accept=".csv" required>
        <br><br>
        <label for="ogg-file">OGG File:</label>
        <input type="file" id="ogg-file" accept=".ogg" required>
    `;
    } else if (option2Radio.checked) {
        uploadContainer.innerHTML = `
        <label for="csv-file">CSV File:</label>
        <input type="file" id="csv-file" accept=".csv" required>
    `;
    }
}

function createTestCase(event) {
    event.preventDefault();

    const testDescriptionInput = document.getElementById('test-description');
    const description = testDescriptionInput.value;

    if (description) {
        // Add your logic here to handle the form submission and create a new test case
        // ...

        closePopup();
    }
}
