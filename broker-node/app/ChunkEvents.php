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
    

	
	public static function addChunkEvent($event_name, $hooknode_id, $session_id, $value){
	
		self::create([

		    'hooknode_id' => $hooknode_id,
		    
		    'session_id' => $session_id,

		    'event_name' => $event_name,

		    'value' => $value

		]);
		
		//$this->save();

	}


}


?>
