<?php

// Class named Database that wraps PDO to run queries to the database
class Database {
    private $pdo;

    public function __construct($dsn, $username, $password) {
        $this->pdo = new PDO($dsn, $username, $password);
    }

    public function query($sql, $params = []) {
        $stmt = $this->pdo->prepare($sql);
        $stmt->execute($params);
        return $stmt->fetchAll(PDO::FETCH_ASSOC);
    }
}

// Create a new Database instance with the required connection details
$db = new Database('mysql:host=localhost;dbname=my_database', 'username', 'password');

// Prepare an SQL query and execute it using the query method
$sql = "SELECT * FROM users WHERE age > :age";
$params = ['age' => 18];
$result = $db->query($sql, $params);

// Print the result
print_r($result);


// Function to render a PHP template file with optional variables
function renderTemplate($templateFile, $variables = []) {
    // Extract variables to be used in the template
    extract($variables);

    // Start output buffering
    ob_start();

    // Include the template file
    include $templateFile;

    // Get the contents of the output buffer
    $output = ob_get_contents();

    // Clean the output buffer and stop buffering
    ob_end_clean();

    // Return the rendered output
    return $output;
}

// Example usage:
$templateFile = 'path/to/template.php';
$variables = ['name' => 'John Doe', 'age' => 30];
echo renderTemplate($templateFile, $variables);

// Create a PHP template file named 'example_template.php'
// with the following content:

/*
<!DOCTYPE html>
<html>
<head>
    <title>User Profile</title>
</head>
<body>
    <h1>User Profile</h1>
    <p>Name: <?= htmlspecialchars($name) ?></p>
    <p>Age: <?= htmlspecialchars($age) ?></p>
</body>
</html>
*/

// Use the 'renderTemplate' function to render the example template
$templateFile = 'path/to/example_template.php';
$variables = ['name' => 'John Doe', 'age' => 30];
echo renderTemplate($templateFile, $variables);