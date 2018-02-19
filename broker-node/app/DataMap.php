<?php

namespace App;

use App\Clients\BrokerNode;
use Illuminate\Database\Eloquent\Model;
use Webpatser\Uuid\Uuid;

class DataMap extends Model
{
    const status = array(
        'unassigned' => 'unassigned',
        'unverified' => 'unverified',
        'complete' => 'complete',
        'error' => 'error',
    );

    /**
     * TODO: Make this a shared trait.
     *  Setup model event hooks
     */
    public static function boot()
    {
        parent::boot();
        self::creating(function ($model) {
            $model->id = (string)Uuid::generate(4);
        });
    }

    protected $table = 'data_maps';
    protected $fillable = [
        'genesis_hash',
        'hash',
        'chunk_idx',
        'address',
        'message',
        'trunkTransaction',
        'branchTransaction',
        'hooknode_id'
    ];
    public $incrementing = false;  // UUID

    /**
     * Builds and loads the datamap in DB given genesis hash and the number of file chunks
     *
     * @param string $genesis_hash A 64 char string
     * @param integer $file_chunk_count the number to generate
     * @return boolean $result response code
     */
    public static function buildMap($genesis_hash, $file_chunk_count)
    {
        //we put the first chunk
        $chunk_idx = 0;
        self::create([
            'genesis_hash' => $genesis_hash,
            'hash' => $genesis_hash,
            'chunk_idx' => $chunk_idx
        ]);

        //it takes the ceiling then disregards any past the correct number
        $hashes_per_db_update = 5;
        $num_groups = ceil(($file_chunk_count) / $hashes_per_db_update);

        //start process of getting hashes
        $hash_generator = self::hashGenerator($genesis_hash, $file_chunk_count);

        for ($i = 0; $i < $num_groups; $i++) {
            $next_group = self::getNextNHashes($hash_generator, $hashes_per_db_update, $chunk_idx);
            $chunk_idx += $hashes_per_db_update;

            foreach ($next_group as $chunk_idx_hash) {

                $curr_hash = $chunk_idx_hash[0];
                $curr_chunk_idx = $chunk_idx_hash[1];

                self::create([
                    'genesis_hash' => $genesis_hash,
                    'hash' => $curr_hash,
                    'chunk_idx' => $curr_chunk_idx
                ]);
            }
        }
    }

    /**
     * Get the next group of hashes in the data map
     *
     * @param generator $hash_generator yields the next hash value in the sequence
     * @param integer $n the number to generate
     * @param $chunk_idx the chunk id to start at
     * @return array Returns array of [chunk id , hash] pairs
     */
    private static function getNextNHashes($hash_generator, $n, $chunk_idx)
    {
        $next_group = array();

        for ($i = 0; $i < $n; $i++) {
            $chunk_idx++;
            $hash_generator->next();
            $next_hash = $hash_generator->current();
            //this is returned if the generator is called too many times
            if ($next_hash == -1) {
                break;
            }

            $next_hash = [$hash_generator->current(), $chunk_idx];
            array_push($next_group, $next_hash);
        }

        return $next_group;
    }

    /**
     * A generator that yields the next of the data map by rehashing the last hash
     * If we call it more than $n times it will just return -1
     *
     * @param string $genesis_hash A 64 char string
     * @param integer $n the number to generate
     * @return User Returns User object or null if not found
     */
    private static function hashGenerator($genesis_hash, $n)
    {
        $hash = $genesis_hash;

        for ($i = 0; $i <= $n; $i++) {
            yield $hash;
            $hash = hash("sha256", $hash);
        }

        while (True) {
            yield -1;
        }
    }

    /**
     * Static Methods.
     */

    public static function getUnassigned()
    {
        return DataMap::where('status', DataMap::status['unassigned'])->get();
    }


    public static function updateChunksPending($chunks)
    {
        //replace with something more efficient
        foreach ($chunks as $chunk) {
            DataMap::where('chunk_idx', $chunk->chunkId)
                ->update([
                    'hooknode_id' => $chunk->hookNodeUrl,
                    'trunkTransaction' => $chunk->trunkTransaction,
                    'branchTransaction' => $chunk->branchTransaction,
                    'status' => self::status['unverified']
                ]);
        }
    }

    /**
     * Instance Methods.
     */

    public function processChunk()
    {
        $brokerReq = (object)[
            "responseAddress" =>
                "{$_SERVER['REMOTE_ADDR']}:{$_SERVER['REMOTE_PORT']}",
            "message" => $this->message,
            "chunkId" => $this->chunk_idx,
            "address" => $this->address,
        ];

        /*TODO
         * may need to return more stuff from processNewChunk
         */

        [$res_type, $updatedChunk] = BrokerNode::processNewChunk($brokerReq);

        switch ($res_type) {
            case 'already_attached':
                // This will be marked as complete in the cron job.
                return $res_type;

            case 'hooknode_unavailable':
                $this->queueForHooknode();
                return $res_type;

            case 'success':
                $this->hooknode_id = $updatedChunk->hookNodeUrl;
                $this->trunkTransaction = $updatedChunk->trunkTransaction;
                $this->branchTransaction = $updatedChunk->branchTransaction;
                $this->status = self::status['unverified'];
                $this->save();

                return $res_type;
        }
    }

    /**
     * Private
     */

    private function queueForHooknode()
    {
        $this->status = DataMap::status['unassigned'];
        $this->save();

        return $this;
    }
}
