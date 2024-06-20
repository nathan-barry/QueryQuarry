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
    fetch("http://localhost:8080/count", {
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
        if (data.sentences != null) {
            for (i=0; i<data.sentences.length; i++) {
                sentences += `${i+1}: ...${data.sentences[i]}...\n\n`
            }
        }
        document.getElementById('resultSentences').innerText = sentences;
    })
    .catch(error => console.error("Error:", error))
}
