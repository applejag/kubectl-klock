Output docs/demo.gif

Require echo
Require tmux
Require kubectl
Require kubectl-klock

Set Shell "bash"
Set TypingSpeed 100ms
Set FontSize 24
Set Width 1024
Set Height 720

Hide
Type "tmux"
Enter
Sleep 3s
Ctrl+B
Type ":set status off"
Enter
Sleep 1s
Show

Sleep 3s

Type "# `kubectl klock pods` is similar to `kubectl get pods -w`"
Sleep 2s
Enter
Sleep 3s

Type "kubectl klock pods"
Sleep 500ms
Enter
Sleep 3s

Ctrl+B
Type '"'
Sleep 5s

Type "# This split represents a different terminal window"
Sleep 500ms 
Enter
Sleep 5s

Type "kubectl run nginx --image nginx"
Sleep 500ms
Enter
Sleep 5s

Type "kubectl run --rm -it ubuntu --image ubuntu"
Sleep 500ms 
Enter
Sleep 3s

Type "echo 'Hello from the ubuntu container'"
Sleep 500ms 
Enter
Sleep 3s

Type "exit"
Sleep 3s
Enter
Sleep 3s

Type "kubectl delete pod nginx"
Sleep 500ms 
Enter

Sleep 10s

Hide
Type "exit"
Enter
Sleep 500ms
Ctrl+C
Sleep 500ms
Type "exit"
Enter
Sleep 500ms