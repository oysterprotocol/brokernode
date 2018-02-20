<?php

namespace Tests\Unit;

use App\HookNode;
use Tests\TestCase;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Foundation\Testing\WithFaker;
use Carbon\Carbon;

class HookNodeTest extends TestCase
{
    use RefreshDatabase;

    public function testgetReadyNode_default()
    {
        self::setupDb();
        $node = HookNode::getNextReadyNode();

        $this->assertEquals($node->ip_address, 'A');
    }

    public function testgetReadyNode_scoreBias()
    {
        self::setupDb();
        $node = HookNode::getNextReadyNode(['time_weight' => 0]);

        $this->assertEquals($node->ip_address, 'A');
    }

    public function testgetReadyNode_timeBias()
    {
        self::setupDb();
        $node = HookNode::getNextReadyNode(['score_weight' => 0]);

        $this->assertEquals($node->ip_address, 'B');
    }

    /**
     * Private
     * */

    private function setupDb() {
        // More recent, high score
        HookNode::create([
            'ip_address' => "A",
            'score' => 500,
            'contacted_at' => Carbon::now()->subSeconds(1),
        ]);

        // Less recent, low score
        HookNode::create([
            'ip_address' => "B",
            'score' => 5,
            'contacted_at' => Carbon::now()->subSeconds(100),
        ]);
    }
}
