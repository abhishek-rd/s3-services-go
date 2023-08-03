let expandedRows = {};
let testCases = [];

window.addEventListener("DOMContentLoaded", () => {
  fetch("/get-json")
    .then((response) => response.json())
    .then((data) => {
      testCases = data.testSuites;
      populateTable(testCases);
    });
});

function populateTable(testSuites) {
  const tableBody = document.querySelector("#testSuitesTable tbody");
  tableBody.innerHTML = "";

  testSuites.forEach((testSuite) => {
    const row = document.createElement("tr");
    const historicalData = testSuite.runs.map((run) => run.tests.filter((test) => test.status === "Passed").length);

    row.innerHTML = `
      <td>${testSuite.id}</td>
      <td>${testSuite.description}</td>
      <td>${getLastRunPassedCount(testSuite)}</td>
      <td>${getLastRunTimestamp(testSuite)}</td>
      <td>${createBarChart(historicalData)}</td>
    `;

    row.dataset.testSuite = JSON.stringify(testSuite); // Store the entire testSuite data as a string

    row.addEventListener("click", () => showRunDetails(testSuite.id));
    tableBody.appendChild(row);
  });
}

function getLastRunPassedCount(testSuite) {
  const lastRun = testSuite.runs[testSuite.runs.length - 1];
  const passedTests = lastRun.tests.filter((test) => test.status === "Passed").length;
  const totalTests = lastRun.tests.length;
  return `${passedTests}/${totalTests}`;
}

function getLastRunTimestamp(testSuite) {
  const lastRun = testSuite.runs[testSuite.runs.length - 1];
  return lastRun.timeStamp;
}

function createBarChart(data) {
  const barWidth = 20;
  let chartHTML = "";
  data.forEach((value) => {
    chartHTML += `<div class="bar" style="height:${value * barWidth}px;"></div>`;
  });
  return `<div class="chart">${chartHTML}</div>`;
}

function showRunDetails(testSuiteId) {
  const tableBody = document.querySelector("#testSuitesTable tbody");
  const rows = tableBody.getElementsByTagName("tr");

  if (expandedRows[testSuiteId]) {
    // If the same test suite row is clicked again, collapse the expanded details
    hideDetailsRow(testSuiteId);
  } else {
    // Show details for the clicked test suite
    hideDetailsRow();
    expandedRows[testSuiteId] = true;

    // Find the test suite data based on the testSuiteId
    const testSuite = findTestSuite(testSuiteId);
    if (testSuite) {
      // Create a new row for the details
      const detailsRow = document.createElement("tr");
      detailsRow.classList.add("details-row");

      const detailsColSpan = rows[0].children.length;
      detailsRow.innerHTML = `
        <td colspan="${detailsColSpan}">
          <h3>Details for Test Suite ${testSuiteId}</h3>
          ${getAllRunsDetails(testSuite)}
        </td>
      `;

      // Insert the new details row right after the clicked row
      for (let i = 0; i < rows.length; i++) {
        const testSuiteData = JSON.parse(rows[i].dataset.testSuite);
        if (testSuiteData.id === testSuiteId) {
          rows[i].insertAdjacentElement("afterend", detailsRow);
          break;
        }
      }
    }
  }
}

function getAllRunsDetails(testSuite) {
  return testSuite.runs.map((run, index) => getRunDetails(testSuite.id, index)).join("");
}

function getRunDetails(testSuiteId, runIndex) {
  const testSuite = findTestSuite(testSuiteId);
  if (testSuite && testSuite.runs.length > runIndex) {
    const run = testSuite.runs[runIndex];
    const tests = run.tests;

    let detailsHTML = `
      <h4>Run ${runIndex + 1} - Timestamp: ${run.timeStamp}</h4>
      <table>
        <thead>
          <tr>
            <th>Description</th>
            <th>Status</th>
            <th>Golden Transcript</th>
            <th>STT Result</th>
            <th>Similarity Score</th>
            <th>File ID</th>
          </tr>
        </thead>
        <tbody>
    `;

    tests.forEach((test) => {
      detailsHTML += `
        <tr>
          <td>${test.description}</td>
          <td>${test.status}</td>
          <td>${test.goldenTranscript}</td>
          <td>${test.STTResult}</td>
          <td>${test.similarityScore}</td>
          <td>${test.fileId}</td>
        </tr>
      `;
    });

    detailsHTML += `
        </tbody>
      </table>
      <hr />
    `;

    return detailsHTML;
  }
  return "";
}

function hideDetailsRow(testSuiteId) {
  const detailsRow = document.querySelector(".details-row");
  if (detailsRow) {
    detailsRow.remove();
  }
  if (testSuiteId) {
    expandedRows[testSuiteId] = false;
  }
}

function findTestSuite(testSuiteId) {
  return testCases.find(testSuite => testSuite.id === testSuiteId);
}

// Function to handle opening the popup
function openPopup() {
  const popup = document.getElementById('popup');
  popup.style.display = 'block';
}

// Function to handle closing the popup
function closePopup() {
  const popup = document.getElementById('popup');
  popup.style.display = 'none';
}

// Function to handle option selection and show appropriate upload boxes
function handleOptionSelection() {
  const option1Radio = document.getElementById('option1');
  const option2Radio = document.getElementById('option2');
  const uploadContainer = document.getElementById('upload-container');

  uploadContainer.innerHTML = '';

  if (option1Radio.checked) {
    uploadContainer.innerHTML = `
      <div>
        <label for="csv-file">CSV File:</label>
        <input type="file" id="csv-file" accept=".csv" required>
      </div>
      <div>
        <label for="ogg-file">OGG File:</label>
        <input type="file" id="ogg-file" accept=".ogg" required>
      </div>
    `;
  } else if (option2Radio.checked) {
    uploadContainer.innerHTML = `
      <div>
        <label for="csv-file">CSV File:</label>
        <input type="file" id="csv-file" accept=".csv" required>
      </div>
    `;
  }
}

// Function to handle form submission and create a new test case
function createTestCase(event) {
  event.preventDefault();

  const testDescriptionInput = document.getElementById('test-description');
  const description = testDescriptionInput.value;

  if (description) {
    const option1Radio = document.getElementById('option1');
    const option2Radio = document.getElementById('option2');
    const csvFileInput = document.getElementById('csv-file');
    const oggFileInput = document.getElementById('ogg-file');

    if (option1Radio.checked) {
      if (!csvFileInput.files[0] || !oggFileInput.files[0]) {
        alert('Please select both CSV and OGG files.');
        return;
      }
      
      const csvFile = csvFileInput.files[0];
      const oggFile = oggFileInput.files[0];

      const formData = new FormData();
      formData.append('description', description);
      formData.append('csvFile', csvFile);
      formData.append('oggFile', oggFile);

      // Call the API endpoint to create a new test case
      fetch('/create-test-case', {
        method: 'POST',
        body: formData,
      }).then(response => {
        if (response.ok) {
          alert('Test case created successfully.');
          closePopup();
          // Refresh the table
          populateTable(testCases);
        } else {
          alert('Error creating test case. Please try again.');
        }
      });

    } else if (option2Radio.checked) {
      if (!csvFileInput.files[0]) {
        alert('Please select a CSV file.');
        return;
      }

      const csvFile = csvFileInput.files[0];

      const formData = new FormData();
      formData.append('description', description);
      formData.append('csvFile', csvFile);

      // Call the API endpoint to create a new test case
      fetch('/create-test-case', {
        method: 'POST',
        body: formData,
      }).then(response => {
        if (response.ok) {
          alert('Test case created successfully.');
          closePopup();
          // Refresh the table
          populateTable(testCases);
        } else {
          alert('Error creating test case. Please try again.');
        }
      });
    }
  } else {
    alert('Please enter a description for the test case.');
  }
}

