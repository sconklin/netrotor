This is a gateway between hamlib's rotctl rotor control libraries, and the network interfaces used by the N1MM ham radio logger.

It's designed to be run on a host that's directly connected to the rotator.

It will poll the rotator using rotctl, and broadcast the position using UDP packets compatible with N1MM.

It will also listen for packets sent by N1MM commanding a movement, and pass those to the rotator using rotctl.

On Ubuntu, UDP is blocked by default by Ubuntu Firewall (ufw). To open ports, do this:

```sudo ufw allow 13010/udp
```

For N1MM, Port 12060 is used for status and 13010 is used for rotor updates

Turn Rotator:
<N1MMRotor>
     <rotor>rotor name</rotor>
     <goazi>55.0</goazi>
     <offset>0.0</offset>
     <bidirectional>0</bidirectional>
     <freqband>14</freqband>   *
</N1MMRotor>

Examples of freqband encoding are 1.8, 3.5, 7, 14, 21, 28

Stop Rotator:
<N1MMRotor>
      <stop>
            <rotor>YaesuCom9</rotor>
            <freqband>21.0</freqband>
      </stop>
</N1MMRotor>

Rotor status update messages sent from the separate N1MM Rotor program on UDP port 13010 are in this format:

rotorname @ rotorheading

They are sent approx every 16 seconds

There are spaces before and after the ‘@’, and the heading is in degrees times ten with no leading zeros, i.e.:
36 degrees - 360
146 degrees - 1460

356 degrees - 3560
