const express = require('express');
const path = require('path');

const app = express();
const PORT = process.env.PORT || 8080;

app.use(express.static(path.join(__dirname, '/all')));


app.get('/', (req, res) => {
    res.sendFile(path.join(__dirname, 'index.html'));
});

app.use((req, res) => {
    res.status(404).sendFile(path.join(__dirname, 'all/404.html'));
});

app.listen(PORT, () => {
    console.log(`Server is running on port ${PORT}`);
});