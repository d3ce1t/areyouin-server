package es.peeple.areyouin.protocol;

import com.google.protobuf.GeneratedMessage;
import java.io.DataInputStream;
import java.io.DataOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;

public class AyiPacket {

    private AyiHeader mHeader;
    private byte[] mData;

    private AyiPacket() {
        mHeader = new AyiHeader();
    }

    public AyiHeader getHeader() {
        return mHeader;
    }

    public byte[] getData() {
        return mData;
    }

    public void writeTo(OutputStream output) throws IOException {

        if (mData.length > 65530)
            throw new IOException("Message exceeds max.size of 65530");

        DataOutputStream dos = new DataOutputStream(output);
        dos.writeByte(mHeader.version);
        dos.writeShort(mHeader.token); // Big-Endian
        dos.writeByte(mHeader.type);
        dos.writeShort(mHeader.size); // Big-Endian
        dos.write(mData);
        dos.flush();
    }

    public static AyiPacket newPacket(byte type, GeneratedMessage message) {
        AyiPacket packet = new AyiPacket();
        packet.mData = message.toByteArray();
        packet.mHeader.type = type;
        packet.mHeader.size = (short) (6 + packet.mData.length);
        return packet;
    }

    public static AyiPacket readFrom(InputStream input) throws IOException {
        AyiPacket packet = new AyiPacket();
        DataInputStream dis = new DataInputStream(input);
        packet.mHeader.version = dis.readByte();
        packet.mHeader.token = dis.readShort();
        packet.mHeader.type = dis.readByte();
        packet.mHeader.size = dis.readShort();
        packet.mData = new byte[packet.mHeader.size - 6];
        dis.read(packet.mData);
        return packet;
    }
}