function updateWordCount() {
    var query = document.getElementById("queryInput").value;
    var wordCount = query.trim() === "" ? 0 : query.trim().split(/\s+/).length;
    document.getElementById("wordCount").innerText = `Number of Words: ${wordCount}`;
}

function sendQuery() {
    var query = document.getElementById("queryInput").value;

    if (query === "") {
        alert("Please enter a query.");
        return;
    }
    fetch("http://localhost:8080/query", {
        method: "POST",
        body: JSON.stringify({
            length: query.length,
            query: query
        }),
        headers: {
            "Content-type": "application/json; charset=UTF-8"
        }
    })
    .then(response => response.json())
    .then(data => {
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
