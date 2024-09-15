function updateWordCount() {
    var queryInput= document.getElementById("queryInput").value;
    var wordCount = queryInput.trim() === "" ? 0 : queryInput.trim().split(/\s+/).length;
    document.getElementById("wordCount").innerText = `Number of Words: ${wordCount}`;
}

function sendQuery() {
    var queryInput= document.getElementById("queryInput").value;
    var datasetSelect = document.getElementById("datasetSelect").value;

    if (queryInput=== "") {
        alert("Please enter a query.");
        return;
    }
    fetch("http://localhost:8081/count", {
        method: "POST",
        body: JSON.stringify({
            dataset: datasetSelect,
            length: queryInput.length,
            query: queryInput
        }),
        headers: {
            "Content-type": "application/json; charset=UTF-8"
        }
    })
    .then(response => response.json())
    .then(data => {
        console.log(data)
        // Display number of occurrences
        document.getElementById('resultNum').innerText = `Number of Occurrences: ${data.occurrences}`;

        // Display sentences for each
        var sentences = ""
        if (data.before != null || data.afters != null) {
            for (i=0; i<data.before.length; i++) {
                sentences += `<b>${i+1}:</b> ...${data.before[i]}<span style="color: #f97316;"><b>${data.query}</b></span>${data.after[i]}...<br><br>`
            }
        }
        document.getElementById('resultSentences').innerHTML = sentences;
    })
    .catch(error => console.error("Error:", error))
}
