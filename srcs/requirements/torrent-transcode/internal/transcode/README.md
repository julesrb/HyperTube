
# simple packaging 

ffmpeg -i the_room.mp4 -c copy -hls_time 5 -hls_list_size 0 -f hls output.m3u8

-i stand for input 
-c for codec (copy means just copy no transcode)
-hls_time for the duration of the chuncks
-hls_list_size 0 to have all the segemtn in the manifest 
-f for format 


This is ok ONLY IF the codec is appropriate 

# encode in 3 different stream of quality

ffmpeg -i the_room.mp4  
-hls_time 5 -hls_list_size 0 -f hls output.m3u8

-filter_complex "[0:v]split=3[v1][v2][v3]; this split into 3 stream

[v1]scale=w=1920*1080[v1out];
[v2]scale=w=1280*720[v2out];
[v3]scale=w=854*480[v3out];

-map "[v1out]" -c:v:0 libx264 -b:v:0 5000k -maxrate:v:0 5350k -bufsize:v:0 7500k 
-map "[v2out]" -c:v:1 libx264 -b:v:1 2800k -maxrate:v:1 2996k -bufsize:v:1 4200k 
-map "[v3out]" -c:v:2 libx264 -b:v:2 1400k -maxrate:v:2 1498k -bufsize:v:2 2100k 

-c:v:0 for code of stream 0

libx264 for a h264 codec

-b:v:0 bitrate rate per sec

-maxrate:v:0 for max bitrate 

-map a:0 -c:a:0 aac -b:a:0 192k -ac 2 ^
-map a:0 -c:a:1 aac -b:a:1 128k -ac 2 ^
-map a:0 -c:a:2 aac -b:a:2 96k -ac 2 ^
-f hls -hls_time 5 -hls_playlist_type vod ^
-hls_flags independent_segments ^
-hls_segment_type mpegts ^
-hls_segment_filename "stream_%v/data%03d.ts" ^
-master_pl_name master.m3u8 ^
-var_stream_map "v:0,a:0 v:1,a:1 v:2,a:2" ^
-hls_list_size 0 ^
stream_%v/playlist.m3u8


//TODO make it scale on smaller size for format that are not 16/9


ffmpeg -i the_room.mp4 \
-filter_complex "[0:v]split=2[v1][v2];[v1]scale=w=1920:h=1080[v1out];[v2]scale=w=1280:h=720[v2out]" \
-map "[v1out]" -c:v:0 libx264 -b:v:0 1750k -maxrate:v:0 2200k -bufsize:v:0 3500k \
-map "[v2out]" -c:v:1 libx264 -b:v:1 600k -maxrate:v:1 100k -bufsize:v:1 1500k \
-map a:0 -c:a:0 aac -b:a:0 192k -ac 2 \
-map a:0 -c:a:1 aac -b:a:1 128k -ac 2 \
-f hls -hls_time 5 -hls_playlist_type vod \
-hls_flags independent_segments \
-hls_segment_type mpegts \
-hls_segment_filename "stream_%v/data%03d.ts" \
-master_pl_name master.m3u8 \
-var_stream_map "v:0,a:0 v:1,a:1" \
-hls_list_size 0 \
stream_%v/playlist.m3u8