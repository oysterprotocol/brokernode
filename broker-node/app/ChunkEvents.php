<?php


namespace App;



use App\Clients\BrokerNode;

use Illuminate\Database\Eloquent\Model;

use Webpatser\Uuid\Uuid;



class ChunkEvents extends Model

{


    public static function boot()
    {

        parent::boot();

        self::creating(function ($model) {

            $model->id = (string)Uuid::generate(4);

        });

    }
    
    protected $table = 'chunk_events';
    protected $fillable = [
        'hook_node_id',
        'session_id',
        'event_name',
        'value'
    ];
    public $incrementing = false;  // UUID
    

	
	public static function addChunkEvent($event_name, $hook_node_id, $session_id, $value){
	
		self::create([

		    'hook_node_id' => $hook_node_id,
		    
		    'session_id' => $session_id,

		    'event_name' => $event_name,

		    'value' => $value

		]);
		
		//$this->save();

	}
	
	//refactor   setting the hook id to be the ip.  
	//TODO:  add session id
	public function addChunkSentToHookNodeEvent($hook_ip){
	    
	    $event_name = "chunk_sent_to_hook";
	     
	    self::addChunkEvent($event_name, $hook_ip, "", "" );
	}
	
	
}


?>