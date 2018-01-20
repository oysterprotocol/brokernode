<?php
namespace Lib;

require_once("Config.php");

use Lib\Config;

class Log {
		
    private $path;
	
	public function __construct() {
		date_default_timezone_set('America/New_York');	
		$this->path  = Config::$log_path;	
	}
	
	public function write($message) {
		$date = new DateTime();
		$log = $this->path . $date->format('Y-m-d').".txt";
		
		if(is_dir($this->path)) {
			if(!file_exists($log)) {
				$fh  = fopen($log, 'a+') or die("Fatal Error !");
				$logContent = "Time : " . $date->format('H:i:s')."\r\n" . $message ."\r\n";
				fwrite($fh, $logContent);
				fclose($fh);
			} else {
				$this->edit($log,$date, $message);
			}
		} else {
			if(mkdir($this->path,0777) === true) {
				$this->write($message);  
			}		
		}
	}

	private function edit($log,$date,$message) {
		$logContent = "Time : " . $date->format('H:i:s')."\r\n" . $message ."\r\n\r\n";
		$logContent = $logContent . file_get_contents($log);
		file_put_contents($log, $logContent);
	}
}