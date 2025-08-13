<?php

class TaskQueue
{
    private $db;

    public function __construct()
    {
        $this->db = new PDO('mysql:host=localhost;dbname=task_database', 'username', 'password');
        
        // SQL statement to create the task_queue table
        $this->db->exec("CREATE TABLE IF NOT EXISTS task_queue (
            id INT AUTO_INCREMENT PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            params TEXT,
            run_at TIMESTAMP NOT NULL,
            status TINYINT NOT NULL DEFAULT 0)"
       );
    }

    public function push($name, $params, $run_at)
    {
        //public function push($task)
        //$stmt = $this->db->prepare("INSERT INTO task_queue (task) VALUES (:task)");
        //$stmt->bindParam(':task', $task);
        
        // Modify the push() method to fill the values in all columns
        $stmt = $this->db->prepare("INSERT INTO task_queue (name, params, run_at, status) VALUES (:name, :params, :run_at, 0)");
        $stmt->bindParam(':name', $name);
        $stmt->bindParam(':params', $params);
        $stmt->bindParam(':run_at', $run_at);
        $stmt->execute();
    }
    
    public function peek()
    {
        // Modify the peek function to only get not executed tasks
        //$stmt = $this->db->query("SELECT task FROM task_queue WHERE status = 0 ORDER BY id ASC LIMIT 1");
        
        // Modify the peek function to only get not executed tasks and run_at is not in the future
        $stmt = $this->db->query("SELECT task FROM task_queue WHERE status = 0 AND run_at <= NOW() ORDER BY id ASC LIMIT 1");
        
        $result = $stmt->fetch(PDO::FETCH_ASSOC);
        return $result['task'];
    }
    
    public function update($taskId)
    {
        $stmt = $this->db->prepare("UPDATE task_queue SET status = 1 WHERE id = :taskId");
        $stmt->bindParam(':taskId', $taskId);
        $stmt->execute();
    }
    
    public function delete($taskId)
    {
        $stmt = $this->db->prepare("DELETE FROM task_queue WHERE id = :taskId");
        $stmt->bindParam(':taskId', $taskId);
        $stmt->execute();
    }
    
    // Method to cleanup tasks with status=1 and created_at is not today
    public function cleanup()
    {
        $stmt = $this->db->prepare("DELETE FROM task_queue WHERE status = 1 AND DATE(created_at) != CURDATE()");
        $stmt->execute();
    }
}