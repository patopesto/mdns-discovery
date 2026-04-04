# VHS script to generate demo gif
# Run from parent directory with 'vhs assets/demo.tape'

Output assets/demo.gif
#Set WindowBar Colorful
Set Width 1920
Set Height 1080
#Set Padding 20
Set PlaybackSpeed 0.5
Set FontSize 15
Env MDNS_FAKE "1"


Type "./mdns-discovery"
Enter
Sleep 5
Down@0.5s 4
Sleep 1
Enter
Sleep 3
Enter
Up@0.5s 3
Sleep 3

Type "q"
Sleep 1

