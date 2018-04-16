This is a gateway between hamlib's rotctl rotor control libraries, and the network interfaces used by the N1MM ham radio logger.

It's designed to be run on a host that's directly connected to the rotator.

It will poll the rotator using rotctl, and broadcast the position using UDP packets compatible with N1MM.

It will also listen for packets sent by N1MM commanding a movement, and pass those to the rotator using rotctl.

