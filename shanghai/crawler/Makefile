.PHONY:	build run run-no-cache test clean

build:
	go build -o shanghai-crawler ./cmd

run:
	go run ./cmd -v daily

run-no-cache:
	go run ./cmd -v daily --no-cache

video:
	ffmpeg -y -f image2 -framerate 4 -i ../data/fig_map_sequences/qgis_covid_map_shanghai%04d.png \
		-vf "fps=10,scale=1000:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
		-loop -1 -final_delay 200 \
		../data/qgis_covid_map_shanghai.gif
	ffmpeg -y -f image2 -framerate 4 -i ../data/fig_map_sequences/qgis_covid_map_shanghai%04d.png \
		-vcodec libx264 -s 1000x1000 \
		-pix_fmt yuv420p \
		../data/qgis_covid_map_shanghai.mp4

test:
	go test -v ./...

clean:
	go clean
	rm -f ./shanghai-crawler
	rm -rf ../data/.web_cache
	rm -rf ../data/fig_map_sequences/*
	rm -rf ../data/*.{gif,mp4}

