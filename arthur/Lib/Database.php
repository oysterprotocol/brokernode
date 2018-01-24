<?php
namespace Lib;

require_once("Log.php");

use Lib\Log;

class Database
{
	private $hostname;
	private $database;
	private $username;
	private $password;
	private $pdo;
	private $sQuery;
	private $isConnected = false;
	private $log;
	private $parameters = array();

	public function __construct() {
		$this->hostname = Config::$hostname;
		$this->database = Config::$database;
		$this->username = Config::$username;
		$this->password = Config::$password;
		$this->log = new Log();	
		$this->connect();
	}
	
	private function connect() {
		$dsn = 'mysql:dbname='.$this->database.';host='.$this->hostname;
		try {
			$this->pdo = new \PDO($dsn, $this->username, $this->password, array(\PDO::MYSQL_ATTR_INIT_COMMAND => "SET NAMES utf8"));
			$this->pdo->setAttribute(\PDO::ATTR_ERRMODE, \PDO::ERRMODE_EXCEPTION);
			$this->pdo->setAttribute(\PDO::ATTR_EMULATE_PREPARES, true);
			$this->isConnected = true;
		} catch (PDOException $e) {
			echo $this->exceptionLog($e->getMessage());
			die();
		}
	}

	private function init($query, $parameters = "") {
		(!$this->isConnected) ? $this->Connect() : "";
		try {
			$this->sQuery = $this->pdo->prepare($query);
			$this->bindMore($parameters);
			if(!empty($this->parameters)) {
				foreach($this->parameters as $param) {
					$parameters = explode("\x7F",$param);
					$this->sQuery->bindParam($parameters[0],$parameters[1]);
				}		
			}
			$this->success = $this->sQuery->execute();		
		} catch(PDOException $e) {
				$this->exceptionLog($e->getMessage(), $query );
		}
	
		$this->parameters = array();
	}		

	public function bind($para, $value) {	
		$this->parameters[count($this->parameters)] = ":" . $para . "\x7F" . utf8_encode($value);
	}
	
	public function bindMore($paramsArray) {
		if(empty($this->parameters) && is_array($paramsArray)) {
			$columns = array_keys($paramsArray);
			foreach($columns as $i => &$column)	{
				$this->bind($column, $paramsArray[$column]);
			}
		}
	}
		
	public function query($query ,$params = array(), $fetchmode = \PDO::FETCH_ASSOC) {
		$query = trim($query);
		$this->init($query, $params);
		$rawStatement = explode(" ", $query);
		$statement = strtolower($rawStatement[0]);
		if ($statement === 'select' || $statement === 'show') {
			return $this->sQuery->fetchAll($fetchmode);
		}
		else if ( $statement === 'insert' ||  $statement === 'update' || $statement === 'delete' ) {
			return $this->sQuery->rowCount();	
		}	
		else {
			return NULL;
		}
	}
		
	public function lastInsertId() {
		return $this->pdo->lastInsertId();
	}	
			
	public function column($query, $params = array()) {
		$this->init($query,$params);
		$columns = $this->sQuery->fetchAll(\PDO::FETCH_NUM);		
		$column = array();
	
		foreach($columns as $cells) {
			$column[] = $cells[0];
		}
	
		return $column;
	}	

	public function row($query, $params = null, $fetchmode = \PDO::FETCH_ASSOC) {				
		$this->init($query, $params);
		return $this->sQuery->fetch($fetchmode);			
	}

	public function single($query, $params = array()) {
		$this->init($query, $params);
		return $this->sQuery->fetchColumn();
	}
     
	private function exceptionLog($message , $sql = "") {
		$exception  = 'Unhandled Exception. <br />';
		$exception .= $message;
		$exception .= "<br /> You can find the error back in the log.";

		if(!empty($sql)) {
			$message .= "\r\nRaw SQL : "  . $sql;
		}
			$this->log->write($message);
		throw new Exception($message);
	}
	
	public function closeConnection() {
		$this->pdo = null;
	}
}