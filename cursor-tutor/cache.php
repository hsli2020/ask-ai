<?php

class Cache {
    private $cache = [];

    public function set($key, $value, $lifetime = 0) {
        $expire = $lifetime > 0 ? time() + $lifetime : 0;
        $this->cache[$key] = ['value' => $value, 'expire' => $expire];
    }

    public function get($key, $default = null) {
        if (isset($this->cache[$key])) {
            if ($this->cache[$key]['expire'] === 0 || $this->cache[$key]['expire'] > time()) {
                return $this->cache[$key]['value'];
            } else {
                $this->delete($key);
            }
        }
        return $default;
    }
    
    public function removeExpiredItems() {
        foreach ($this->cache as $key => $item) {
            if ($item['expire'] !== 0 && $item['expire'] <= time()) {
                $this->delete($key);
            }
        }
    }
    
    public function delete($key) {
        unset($this->cache[$key]);
    }

    public function clear() {
        $this->cache = [];
    }
}