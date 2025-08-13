<?php

class DependencyInjectionContainer
{
    private $services = [];
    private $instances = [];

    public function set($name, callable $service, $singleton = true)
    {
        $this->services[$name] = $service;
        if ($singleton) {
            $this->instances[$name] = null;
        }
    }

    public function get($name)
    {
        if (!isset($this->services[$name])) {
            throw new Exception("Service not found: " . $name);
        }

        if (isset($this->instances[$name])) {
            if ($this->instances[$name] === null) {
                $this->instances[$name] = $this->services[$name]($this);
            }
            return $this->instances[$name];
        }

        return $this->services[$name]($this);
    }
}

// Define UserService and VideoService classes
class UserService {}
class VideoService {}

// Modify ExampleService to accept UserService and VideoService as dependencies
class ExampleService
{
    private $userService;
    private $videoService;

    public function __construct(UserService $userService, VideoService $videoService)
    {
        $this->userService = $userService;
        $this->videoService = $videoService;
    }
}

// Create a new instance of the DependencyInjectionContainer
$container = new DependencyInjectionContainer();

// Define services for UserService and VideoService
$container->set('userService',  function ($container) { return new UserService(); });
$container->set('videoService', function ($container) { return new VideoService(); });

// Update the exampleService definition to include the dependencies
$container->set('exampleService', function ($container) {
    return new ExampleService(
        $container->get('userService'),
        $container->get('videoService')
    );
});

// Retrieve the "exampleService" instance from the container
$exampleService = $container->get('exampleService');

// Use the "exampleService" instance
$exampleService->doSomething();