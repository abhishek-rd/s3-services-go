let expandedRows = {};

window.addEventListener("DOMContentLoaded", () => {
  fetch("/get-json")
    .then((response) => response.json())
    .then((data) => {
      populateTable(data.testSuites);
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


// function createBarChart(data) {
//   const maxBarHeight = 100; // Maximum height of the bar
//   let chartHTML = "";
//   data.forEach((value, index) => {
//     const totalTests = value + (data[index] - value); // Calculate total tests
//     const passPercentage = (value / totalTests) * 100;
//     let barColor;

//     if (passPercentage === 100) {
//       barColor = "#4caf50"; // Green
//     } else if (passPercentage < 25) {
//       barColor = "#f44336"; // Red
//     } else {
//       barColor = "#ffc107"; // Yellow
//     }

//     const barHeight = (totalTests === 0) ? 0 : (passPercentage / 100) * maxBarHeight;
//     chartHTML += `
//       <div class="bar" style="background-color: ${barColor};">
//         <div class="bar-fill" style="height:${barHeight}px;"></div>
//         <div class="bar-label">${value}/${totalTests}</div>
//       </div>
//     `;
//   });
//   return `<div class="chart">${chartHTML}</div>`;
// }



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
  const tableBody = document.querySelector("#testSuitesTable tbody");
  const rows = tableBody.getElementsByTagName("tr");

  for (let i = 0; i < rows.length; i++) {
    const testSuiteData = JSON.parse(rows[i].dataset.testSuite);
    if (testSuiteData.id === testSuiteId) {
      return testSuiteData;
    }
  }

  return null;
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
      const csvFileInput = document.createElement('input');
      csvFileInput.type = 'file';
      csvFileInput.id = 'csv-file';
      csvFileInput.accept = '.csv';
      csvFileInput.required = true;
  
      const oggFileInput = document.createElement('input');
      oggFileInput.type = 'file';
      oggFileInput.id = 'ogg-file';
      oggFileInput.accept = '.ogg';
      oggFileInput.required = true;
  
      uploadContainer.appendChild(csvFileInput);
      uploadContainer.appendChild(document.createElement('br'));
      uploadContainer.appendChild(document.createElement('br'));
      uploadContainer.appendChild(oggFileInput);
    } else if (option2Radio.checked) {
      const csvFileInput = document.createElement('input');
      csvFileInput.type = 'file';
      csvFileInput.id = 'csv-file';
      csvFileInput.accept = '.csv';
      csvFileInput.required = true;
  
      uploadContainer.appendChild(csvFileInput);
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
  
      let uploadContainerHTML = '';
  
      if (option1Radio.checked) {
        uploadContainerHTML = `
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
        uploadContainerHTML = `
          <div>
            <label for="csv-file">CSV File:</label>
            <input type="file" id="csv-file" accept=".csv" required>
          </div>
        `;
      }
  
      const uploadContainer = document.getElementById('upload-container');
      uploadContainer.innerHTML = uploadContainerHTML;
  
      // Generate a unique ID for the new test case (you can modify this logic if needed)
      const id = testCases.length > 0 ? testCases[testCases.length - 1].id + 1 : 1;
  
      // Generate a URL for the new test case
      const url = `testcase${id}.html`;
  
      // Create the new test case object with an empty tests array and the generated URL
      const newTestCase = { id, description, url, tests: [] };
  
      // Add the new test case to the testCases array
      testCases.push(newTestCase);
  
      // Clear the input field
      testDescriptionInput.value = '';
  
      // Render the updated test cases
      renderTestCases();
  
      // Close the popup
      closePopup();
    }
  }
  
// Event listener for the "+" button
const addTestCaseButton = document.getElementById('add-test-case-button');
if (addTestCaseButton) {
  addTestCaseButton.addEventListener('click', openPopup);
}

// Event listener for the close button in the popup
const closeButton = document.querySelector('.popup-content .close-button');
if (closeButton) {
  closeButton.addEventListener('click', closePopup);
}

// Event listener for option selection
const option1Radio = document.getElementById('option1');
const option2Radio = document.getElementById('option2');
if (option1Radio && option2Radio) {
  option1Radio.addEventListener('change', handleOptionSelection);
  option2Radio.addEventListener('change', handleOptionSelection);
}

// Event listener for form submission in the popup
const newTestCaseForm = document.getElementById('new-test-case-form');
if (newTestCaseForm) {
  newTestCaseForm.addEventListener('submit', createTestCase);
}
