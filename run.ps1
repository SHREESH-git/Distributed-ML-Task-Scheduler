# Helper script to execute Go commands with the correct PATH

$goBinPath = "C:\Program Files\Go\bin"
if (($env:Path -split ';') -notcontains $goBinPath) {
    $env:Path = "$goBinPath;$env:Path"
}

# Add convenient shorthands to run master and worker
if ($args.Count -eq 1 -and $args[0] -eq "master") {
    go run master.go
} elseif ($args.Count -eq 1 -and $args[0] -eq "worker") {
    go run worker.go
} else {
    # Execute the go command with all passed arguments
    go @args
}
