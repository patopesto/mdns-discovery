# VHS script to generate demo gif
# Run from parent directory with 'vhs assets/demo.tape'

Output assets/demo.gif
#Set WindowBar Colorful
Set Width 1920
Set Height 1080
#Set Padding 20
Set PlaybackSpeed 0.5
Set FontSize 15

Hide
Type "go build ."
Enter
Type "clear"
Enter
Show


Type "./mdns-discovery"
Enter
Sleep 5
Down@0.5s 5
Sleep 1
Up@0.5s 3
Sleep 3

Type "q"
Sleep 1


Hide
Type "rm ./mdns-discovery"
Enter
