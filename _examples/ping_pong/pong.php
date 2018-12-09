<?php
while(!feof(STDIN)){
    $line = fgets(STDIN);
    $req=$line;
    
    $req=str_replace("\\n","\n",trim($line));
    
    $resp=sprintf("pong:%s",$req);
    
    echo str_replace("\n","\\n",$resp);
    echo "\n";
}