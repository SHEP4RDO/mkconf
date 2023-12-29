# Module `mkconf`

The `mkconf` module is a tool for managing configuration files in an application. It provides loading, monitoring and updating of configurations, allowing easy integration of configuration changes without restarting the application.

## Key Features

### 1. Flexibility in configuration format selection

The module supports various configuration file formats, including JSON, YAML, XML, TOML and INI. You can easily select the right format for your application or add your own.

### 2. Automatic change monitoring

`mkconf` provides automatic monitoring of changes in configuration files. When changes are detected, a change is signaled. When receiving the signal, you can determine the logic of reloading the configuration yourself without having to stop the application itself.

### 3. Updating configuration without restarting

You can update the configuration in rantime, applying the changes without restarting the application. This is useful for scenarios where dynamic configuration changes are required.

### 4. Support for multiple configurations

The module supports working with multiple configurations at the same time. You can easily add, delete and update configurations in your application.

### 5. Change tracking

For change auditing, `mkconf` provides tracking and logging of configuration changes. This helps in debugging and understanding what parameters were changed and when.

### 6. Multithreading and safety

The module provides multi-threaded processing and security while reading and writing configurations. This is important to prevent conflicts during simultaneous access from different parts of the application.

## Supported formats

-   JSON
-   YAML
-   XML
-   TOML
-   INI

## Usage

Instructions on how to use the `mkconf` module can be found in the corresponding [wiki](https://github.com/SHEP4RDO/mkconf/wiki) of the project.

## License:

`mkconf` is distributed under the MIT license. More information can be found in the LICENSE file.