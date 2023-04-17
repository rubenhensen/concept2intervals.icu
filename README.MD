# Concept 2 logbook to intervals.icu
Very simple script. Every 24 hours it downloads all activities from Concept2 logbook and uploads them to intervals.icu.

To run use my hub.docker image or build your own using docker build, then:
`docker run hebury/c2i -e "conceptkey=yourConcept2APIkeyHere" -e "intervalsid=yourIntervals.icuIdHere" -e "intervalskey=yourIntervals.icuAPIkeyHere"`