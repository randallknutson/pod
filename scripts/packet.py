
from Crypto.Cipher import AES
from CryptoMobile.Milenage import Milenage

from textwrap import wrap
import sys
import argparse

# noqa: E501


# https://tools.ietf.org/html/rfc3748
CODE_BIN = {
    1: 'Request',
    2: 'Response',
    3: 'Success',
    4: 'Failure'

}

SUBTYPE = {
    1: "AKA-Challenge",
    2: "AKA-Authentication-Reject",
    4: "AKA-Synchronization-Failure",
    5: "AKA-Identity",
    10: "SIM-Start",
    11: "SIM-Challenge",
    12: "AKA-Notification and SIM-Notification",
    13: "AKA-Reauthentication and SIM-Reauthentication",
    14: "AKA-Client-Error and SIM-Client-Error",
}

ATTRIBUTE_TYPE = {
    1: "AT_RAND",
    2: "AT_AUTN",
    3: "AT_RES",
    4: "AT_AUTS",
    6: "AT_PADDING",
    7: "AT_NONCE_MT",
    10: "AT_PERMANENT_ID_REQ",
    11: "AT_MAC",
    12: "AT_NOTIFICATION",
    13: "AT_ANY_ID_REQ",
    14: "AT_IDENTITY",
    15: "AT_VERSION_LIST",
    16: "AT_SELECTED_VERSION",
    17: "AT_FULLAUTH_ID_REQ",
    19: "AT_COUNTER",
    20: "AT_COUNTER_TOO_SMALL",
    21: "AT_NONCE_S",
    22: "AT_CLIENT_ERROR_CODE",
    126: "???_CONTROLER_OR_NODE_IV??",
    129: "AT_IV",
    130: "AT_ENCR_DATA",
    132: "AT_NEXT_PSEUDONYM",
    133: "AT_NEXT_REAUTH_ID",
    134: "AT_CHECKCODE",
    135: "AT_RESULT_IND",
}


class Ble():
    def __init__(self, input=None, bin=None, eap=False):
        if not bin:
            hex = input.split(',')
            bin = list(map(lambda x: int(x, 16), hex))
        self._from = bin[8:12]
        self._to = bin[12:16]
        self.preamble = bin[0]
        self.unknown = bin[1:7]
        self.crc = bin[-3:]
        self.data = bin[16:]
        self.eap = None
        if eap:
            self.eap = Eap(bin=self.data)

    def __str__(self):
        ret = f"""
            Preamble: {self.preamble}
            From: {','.join(list(map(to_hex,self._from)))}
            To: {','.join(list(map(to_hex,self._to)))}
            Data: {self.data} :: {len(self.data)}
            Data Hex: {','.join(list(map(to_hex, self.data)))}
            Unknown: {self.unknown}
            UnknownHex: {','.join(list(map(to_hex, self.unknown)))}
        """
        if self.eap:
            ret += str(self.eap)
        return ret


class Attribute():
    def __init__(self, type, data):
        self.type_bin = type
        self.type = ATTRIBUTE_TYPE.get(self.type_bin, "not documented")
        self.data = data

    def __str__(self):
        return f"""
                Type: {self.type} :: {self.type_bin}
                Data: {self.data} :: {len(self.data)}
                Hex: {','.join(list(map(to_hex, self.data)))}
        """


class Res(Attribute):
    def __init__(self, data):
        self.type_bin = 3
        self.type = ATTRIBUTE_TYPE.get(self.type_bin, "not documented")
        self.length = (data[0] * 16 + data[1]) / 8
        self.data = data[2:]

    def __str__(self):
        return f"""
                Type: {self.type} :: {self.type_bin}
                Data: {self.data}
                Length: {self.length}
                Hex: {','.join(list(map(to_hex, self.data)))}
        """


DECODED_ATTRIBUTES = {
    3: Res,
}


class Eap():

    def __init__(self, input=None, bin=None):
        """
        Example input
            01,48,00,38,17,01,00,00,02,05,00,00,5d,2a,6e,39,dc,59,b9,b9,04,cd,89,cc,42,3c,c1,53,01,05,00,00,89,ce,f1,27,62,ee,df,fe,25,d6,41,90,b7,16,50,c9,7e,02,00,00,cb,07,56,3b,
        """
        if not bin:
            hex = input.split(',')
            bin = list(map(lambda x: int(x, 16), hex))

        self.length = bin[2] * 16 + bin[3]
        self.code_bin = bin[0]
        self.code = CODE_BIN[self.code_bin]
        self.identifier = bin[1]
        self.type = None
        if self.length <= 4:
            return
        self.type = str(bin[4]) + "(should be 23)"
        self.subtype_int = bin[5]
        self.subtype = SUBTYPE[self.subtype_int]
        self.attributes = []
        tail = bin[8:]
        while tail:
            attribute_type = tail[0]
            length = tail[1] * 4  # multiple of 4 bytes
            attribute_data = tail[2:length]

            if attribute_type in [1, 2]:  # remove 2 reserved bytes
                assert(attribute_data[0] == 0)
                assert(attribute_data[1] == 0)

                attribute_data = attribute_data[2:]

            attribute = Attribute(attribute_type, attribute_data)
            if attribute_type in DECODED_ATTRIBUTES:
                attribute = DECODED_ATTRIBUTES[attribute_type](attribute_data)

            self.attributes.append(attribute)
            tail = tail[length:]

    def __str__(self):
        ret = f"""
        Length: {self.length}
        Code:: {self.code}
        Identifier: {self.identifier}
        """
        if self.type is not None:
            ret += f"""
                Type: {self.type}
                Subtype: {self.subtype} :: {self.subtype_int}
                Attributes: {len(self.attributes)}
            """
            for i in self.attributes:
                ret += str(i)
        return ret


def to_hex(a):
    return "{0:02x}".format(a)


class PodCommand():
    def __init__(self, cmd):
        self.cmd = cmd
        self.data = list(map(lambda x: int(x, 16), wrap(cmd, 2)))

    def __str__(self):
        return f"""
            Raw: {self.cmd}
            Data: {self.data}
            Hex: {','.join(list(map(to_hex, self.data)))}
        """


def decrypt(args):

    # 2020-03-04 19:37:05.044  1342  1401 I TwiBleManager bytes received on ble 54,57,11,a1,05,08,04,a0,08,20,2e,a9,08,20,2e,a8,6d,fd,d5,e9,26,6d,54,9e,82,0e,a9,a2,68,0c,8a,88,18,0f,d3,df,34,2a,13,e8,8e,cd,3a,db,4f,a0,95,eb,0a,ed,c1,e0,e8,b6,c9,48,07,8e,d0,c9,72,
    # 2020-03-04 19:37:05.044  1342  1401 I CentralClient  notifyReadCompleted 0
    # 2020-03-04 19:37:05.045  1342  1401 D EnDecryptionModule Decrypt NONCE is 6c,ff,5d,18,b7,61,6c,ae,80,00,00,00,02,
    # 2020-03-04 19:37:05.045  1342  1401 I TwiCommunicationProtocol ------------On read  notification number is: 2
    # 2020-03-04 19:37:05.046  1342  1401 I TwiCommunicationProtocol 3b decrypted bytes => 30,2e,30,3d,00,1f,ff,ff,ff,ff,30,17,01,15,03,1d,00,08,08,00,04,02,08,13,9a,51,00,11,92,91,00,ff,ff,ff,ff,83,71,

    # Decrypt
    ble = Ble("54,57,11,a1,05,08,04,a0,08,20,2e,a9,08,20,2e,a8,6d,fd,d5,e9,26,6d,54,9e,82,0e,a9,a2,68,0c,8a,88,18,0f,d3,df,34,2a,13,e8,8e,cd,3a,db,4f,a0,95,eb,0a,ed,c1,e0,e8,b6,c9,48,07,8e,d0,c9,72")
    received_hex = ''.join(list(map(to_hex, ble.data)))
    nonce = "6c,ff,5d,18,b7,61,6c,ae,80,00,00,00,02".replace(',', '')
    expected = "30,2e,30,3d,00,1f,ff,ff,ff,ff,30,17,01,15,03,1d,00,08,08,00,04,02,08,13,9a,51,00,11,92,91,00,ff,ff,ff,ff,83,71"
    data = received_hex
    ck = "55,79,9f,d2,66,64,cb,f6,e4,76,52,5e,2d,ee,52,c6".replace(',', '')

    print(f"Decrypt: CK: {ck}, Nonce: {nonce}, Data:{data}")
    bin_nonce = bytes.fromhex(nonce)
    bin_ck = bytes.fromhex(ck)
    bin_data = bytes.fromhex(data)

    print("Nonce", bin_nonce, len(bin_nonce))
    print("Ck", bin_ck, len(bin_ck))
    print("Data", bin_data, len(bin_data))

    cipher = AES.new(bin_ck, AES.MODE_CCM, bin_nonce)
    # cipher.update(bin_data)
    decrypted = cipher.decrypt(bin_data)

    print("Decrypted: ", decrypted.hex(','), len(decrypted))
    print("Expected:  ", expected, len(expected.replace(',', '')) / 2)


def encrypt(args):
    # 2020-03-04 19:37:04.709  1342  1401 I TwiBleManager bytes going on ble 54,57,11,01,07,00,03,40,08,20,2e,a8,08,20,2e,a9,ab,35,d8,31,60,9b,b8,fe,3a,3b,de,5b,18,37,24,9a,16,db,f8,e4,d3,05,e9,75,dc,81,7c,37,07,cc,41,5f,af,8a,
    # Encrypt
    # 2020-03-04 19:37:04.737  1342  1401 I PodComm pod command: FFFFFFFF2C060704FFFFFFFF817A
    # 2020-03-04 19:37:04.709  1342  1401 I TwiBleManager bytes going on ble 54,57,11,01,07,00,03,40,08,20,2e,a8,08,20,2e,a9,ab,35,d8,31,60,9b,b8,fe,3a,3b,de,5b,18,37,24,9a,16,db,f8,e4,d3,05,e9,75,dc,81,7c,37,07,cc,41,5f,af,8a,
    # 2020-03-04 19:37:04.692  1342  1401 D EnDecryptionModule Encrypt NONCE is 6c,ff,5d,18,b7,61,6c,ae,00,00,00,00,01,
    ble = Ble("54,57,11,01,07,00,03,40,08,20,2e,a8,08,20,2e,a9,ab,35,d8,31,60,9b,b8,fe,3a,3b,de,5b,18,37,24,9a,16,db,f8,e4,d3,05,e9,75,dc,81,7c,37,07,cc,41,5f,af,8a")
    ble_hex = ''.join(list(map(to_hex, ble.data)))
    expected = bytes.fromhex("FFFFFFFF2C060704FFFFFFFF817A").hex(',')
    nonce = "6c,ff,5d,18,b7,61,6c,ae,00,00,00,00,01".replace(',', '')
    ck = "55,79,9f,d2,66,64,cb,f6,e4,76,52,5e,2d,ee,52,c6".replace(',', '')
    data = ble_hex
    print(f"Encrypt: CK: {ck}, Nonce: {nonce}, Data:{data}")
    bin_nonce = bytes.fromhex(nonce)
    bin_ck = bytes.fromhex(ck)
    bin_data = bytes.fromhex(data)

    print("Nonce", bin_nonce, len(bin_nonce))
    print("Ck", bin_ck, len(bin_ck))
    print("Data", bin_data, len(bin_data))

    cipher = AES.new(bin_ck, AES.MODE_CCM, bin_nonce)
    # cipher.update(bin_data)
    decrypted = cipher.decrypt(bin_data)

    print("Decrypted: ", decrypted.hex(','), len(decrypted))
    print("Expected:  ", expected, len(expected.replace(',', '')) / 2)


def milenage(ltk, rand, seq, ck, want_res, autn):
    bin_ltk = bytes.fromhex(ltk)
    bin_rand = bytes.fromhex(rand)
    bin_seq = bytes.fromhex(seq)
    bin_ck = bytes.fromhex(ck)
    bin_autn = bytes.fromhex(autn)
    print("LTK: ", bin_ltk.hex(), len(bin_ltk))
    print("Autn: ", bin_autn.hex(","), len(bin_autn))
    print("Rand: ", bin_rand.hex(), len(bin_rand))
    print("Seq: ", bin_seq.hex(","), len(bin_seq))
    bin_amf = bin_autn[6:8]
    print("Amf: ", bin_amf.hex(), len(bin_amf))
    # 0xf6203e12d502c2cd
    # 0x18b32cc76a676d2b

    op = 'f6203e12d502c2cd18b32cc76a676d2b'
    # op = '18b32cc76a676d2bf6203e12d502c2cd'
    op = 'cdc202d5123e20f62b6d676ac72cb318'
    bin_op = bytes.fromhex(op)
    print("Op: ", bin_op.hex(","), len(bin_op))

    m = Milenage(bin_op)
    res, res_ck, ik, ak = m.f2345(bin_ltk, bin_rand)

    print("CK is: ", bin_ck.hex(","), len(bin_ck))
    print("Received CK is: ", res_ck.hex(","), len(res_ck))
    print("Got  Res is: ", res.hex(","), len(res))
    print("Want Res is: ", want_res, len(want_res) / 2)

    print("IK is: ", ik.hex(","), len(ik))
    print("AK is: ", ak.hex(","), len(ak))


def _xor(a, b):
    ret = bytearray(16)
    for i in range(len(a)):
        ret[i] = a[i] ^ b[i]
    return ret


def _get_milenage(opc, k, rand, sqn, amf):
    '''
        Computes milenage values from OPc, K, RAND, SQN and AMF
        Returns a concatenated list (RAND + AUTN + IK + CK + RES) that
        will be sent back as the response to the client (hostapd). This
        is a python re-write of the function eap_aka_get_milenage() from
        src/simutil.c
    '''
    opc = bytearray.fromhex(opc)
    k = bytearray.fromhex(k)
    # rand gets returned, so it should be left as a hex string
    _rand = bytearray.fromhex(rand)
    sqn = bytearray.fromhex(sqn)
    amf = bytearray.fromhex(amf)
    aes1 = AES.new(bytes(k), AES.MODE_ECB)
    tmp1 = _xor(_rand, opc)
    tmp1 = aes1.encrypt(bytes(tmp1))
    tmp1 = bytearray(tmp1)
    tmp2 = bytearray()
    tmp2[0:6] = sqn
    tmp2[6:2] = amf
    tmp2[9:6] = sqn
    tmp2[15:2] = amf
    tmp3 = bytearray(16)
    for i in range(len(tmp1)):
        tmp3[(i + 8) % 16] = tmp2[i] ^ opc[i]
    tmp3 = _xor(tmp3, tmp1)
    aes2 = AES.new(bytes(k), AES.MODE_ECB)
    tmp1 = aes2.encrypt(bytes(tmp3))
    tmp1 = bytearray(tmp1)
    tmp1 = _xor(tmp1, opc)
    maca = _bytetostring(tmp1[0:8])
    tmp1 = _xor(_rand, opc)
    aes3 = AES.new(bytes(k), AES.MODE_ECB)
    tmp2 = aes3.encrypt(bytes(tmp1))
    tmp2 = bytearray(tmp2)
    tmp1 = _xor(tmp2, opc)
    tmp1[15] ^= 1
    aes4 = AES.new(bytes(k), AES.MODE_ECB)
    tmp3 = aes4.encrypt(bytes(tmp1))
    tmp3 = bytearray(tmp3)
    tmp3 = _xor(tmp3, opc)
    res = _bytetostring(tmp3[8:16])
    ak = _bytetostring(tmp3[0:6])
    for i in range(len(tmp1)):
        tmp1[(i + 12) % 16] = tmp2[i] ^ opc[i]
    tmp1[15] ^= 1 << 1
    aes5 = AES.new(bytes(k), AES.MODE_ECB)
    tmp1 = aes5.encrypt(bytes(tmp1))
    tmp1 = bytearray(tmp1)
    tmp1 = _xor(tmp1, opc)
    ck = _bytetostring(tmp1)
    for i in range(len(tmp1)):
        tmp1[(i + 8) % 16] = tmp2[i] ^ opc[i]
    tmp1[15] ^= 1 << 2
    aes6 = AES.new(bytes(k), AES.MODE_ECB)
    tmp1 = aes6.encrypt(bytes(tmp1))
    tmp1 = bytearray(tmp1)
    tmp1 = _xor(tmp1, opc)
    ik = _bytetostring(tmp1)
    tmp1 = bytearray.fromhex(ak)
    autn = bytearray(6)
    for i in range(0, 6):
        autn[i] = sqn[i] ^ tmp1[i]
    autn[6:2] = amf
    autn[8:8] = bytearray.fromhex(maca)[0:8]
    autn = _bytetostring(autn)
    return rand, autn, ik, ck, res


def _bytetostring(b):
    return ''.join(format(x, '02x') for x in b)


def key_exchange(args):
    # SP1, SP2
    # 53,50,31,3d,00,04,08,20,2e,a9,2c,53,50,32,3d,00,0b,08,20,2e,ab,2c,03,0e,01,00,01,91,
    # SP1 -> pod ID
    # SP2 -> get pod status command

    # SPS1
    # 53,50,53,31,3d,00,30,38,82,8b,3d,0c,24,4d,ce,d0,b2,a2,e5,9c,69,93,3d,86,56,32,e5,b5,f2,8b,df,c2,15,24,d0,1f,d6,51,7d,54,b3,82,9f,96,8c,d6,6d,10,8f,44,0b,c6,0b,89,ee,
    sps1 = "38,82,8b,3d,0c,24,4d,ce,d0,b2,a2,e5,9c,69,93,3d,86,56,32,e5,b5,f2,8b,df,c2,15,24,d0,1f,d6,51,7d,54,b3,82,9f,96,8c,d6,6d,10,8f,44,0b,c6,0b,89,ee".split(',')
    pdm_key = sps1[:32]
    pdm_nonce = sps1[32:]
    # received from pod
    # 53,50,53,31,3d,00,30,1d,2d,d1,71,72,aa,dc,c9,fb,91,51,e0,ed,ad,67,e5,01,b2,86,00,de,81,f7,5b,7b,b0,9d,5d,cb,1d,00,5f,d5,2d,a5,2f,1d,6e,0b,5c,60,bc,fb,35,58,08,74,51,
    sps1_pod = "1d,2d,d1,71,72,aa,dc,c9,fb,91,51,e0,ed,ad,67,e5,01,b2,86,00,de,81,f7,5b,7b,b0,9d,5d,cb,1d,00,5f,d5,2d,a5,2f,1d,6e,0b,5c,60,bc,fb,35,58,08,74,51".split(',')
    pod_key = sps1_pod[:32]
    pod_nonce = sps1_pod[32:]

    pdm_sps2 = "37,a0,da,ac,48,75,43,a4,26,eb,b3,a8,00,8c,09,f5".replace(',', '')
    pod_sps2 = "19,3b,10,dc,3f,8c,c0,6a,60,49,c4,0e,0a,43,ab,9a".replace(',', '')
    ltk = "c0,77,28,99,72,09,72,a3,14,f5,57,de,66,d5,71,dd".replace(',', '')

    pdm_key_bin = bytes.fromhex(pdm_key)
    pdm_nonce_bin = bytes.fromhex(pdm_nonce)
    pdm_sps2_bin = bytes.fromhex(pdm_sps2)

    pod_key_bin = bytes.fromhex(pod_key)
    pod_nonce_bin = bytes.fromhex(pod_nonce)
    pod_sps2_bin = bytes.fromhex(pod_sps2)

    ltk_bin = bytes.fromhex(ltk)

    print("PDM Key", pdm_key_bin.hex(), len(pdm_key_bin))
    print("PDM Nonce", pdm_nonce_bin.hex(), len(pdm_nonce_bin))
    print("PDM SPS2", pdm_sps2_bin.hex(), len(pdm_sps2_bin))

    print("POD Key", pod_key_bin.hex())
    print("POD Nonce", pod_nonce_bin.hex())
    print("POD SPS2", pod_sps2_bin.hex())

    print("LTK", ltk_bin.hex(), len(ltk_bin))


def eap_aka(args):
    # 2020-03-04 19:37:04.014  1342  2618 D PairingMasterModule LTK: c0,77,28,99,72,09,72,a3,14,f5,57,de,66,d5,71,dd,
    # 2020-03-04 19:37:04.086  1342  2618 I ConnectionManager **************** MY ID IS *********************08,20,2e,a8,
    # 2020-03-04 19:37:04.087  1342  2618 D EapAkaMasterModule Eap aka start session sequence 00,00,00,00,00,01,

    # 2020-03-04 19:37:04.088  1342  2618 D SecurityMasterManager send this msg on comm  01,bd,00,38,17,01,00,00,02,05,00,00,00,c5,5c,78,e8,d3,b9,b9,e9,35,86,0a,72,59,f6,c0,01,05,00,00,c2,cd,12,48,45,11,03,bd,77,a6,c7,ef,88,c4,41,ba,7e,02,00,00,6c,ff,5d,18,
    # print(Ble("54,57,10,02,05,00,07,00,08,20,2e,a8,08,20,2e,a9,01,bd,00,38,17,01,00,00,02,05,00,00,00,c5,5c,78,e8,d3,b9,b9,e9,35,86,0a,72,59,f6,c0,01,05,00,00,c2,cd,12,48,45,11,03,bd,77,a6,c7,ef,88,c4,41,ba,7e,02,00,00,6c,ff,5d,18", eap=True))
    print(Eap("01,bd,00,38,17,01,00,00,02,05,00,00,00,c5,5c,78,e8,d3,b9,b9,e9,35,86,0a,72,59,f6,c0,01,05,00,00,c2,cd,12,48,45,11,03,bd,77,a6,c7,ef,88,c4,41,ba,7e,02,00,00,6c,ff,5d,18"))
    # print(Eap("03,be,00,0 4"))

    #  Mu ID in construct packet connection 08,20,2e,a8
    # 2020-03-04 19:37:04.408  1342  1402 D SecurityMasterManager received this msg on comm  02,bd,00,1c,17,01,00,00,03,03,00,40,a4,0b,c6,d1,38,61,44,7e,7e,02,00,00,b7,61,6c,ae,
    print(Eap("02,bd,00,1c,17,01,00,00,03,03,00,40,a4,0b,c6,d1,38,61,44,7e,7e,02,00,00,b7,61,6c,ae"))
    # 2020-03-04 19:37:04.529  1342  1401 D EapAkaMasterModule Sequence number after session start is 00,00,00,00,00,02,
    # 2020-03-04 19:37:04.540  1342  1401 D SecurityMasterManager Controller CK = 55,79,9f,d2,66,64,cb,f6,e4,76,52,5e,2d,ee,52,c6,
    # 2020-03-04 19:37:04.541  1342  1401 D SecurityMasterManager Controller IV = 6c,ff,5d,18,
    # 2020-03-04 19:37:04.541  1342  1401 D SecurityMasterManager Node IV = b7,61,6c,ae,
    ck = "55,79,9f,d2,66,64,cb,f6,e4,76,52,5e,2d,ee,52,c6".replace(',', '')

    rand = "c2,cd,12,48,45,11,03,bd,77,a6,c7,ef,88,c4,41,ba".replace(",", "")
    seq = "00,00,00,00,00,01".replace(",", "")
    ltk = "c0,77,28,99,72,09,72,a3,14,f5,57,de,66,d5,71,dd".replace(",", "")
    res = "a4,0b,c6,d1,38,61,44,7e"
    autn = "00,c5,5c,78,e8,d3,b9,b9,e9,35,86,0a,72,59,f6,c0".replace(",", "")
    milenage(ltk, rand, seq, ck, res, autn)
    print("-----------")
    amf = "b9b9"
    rand, autn, ik, ck, res = _get_milenage('cdc202d5123e20f62b6d676ac72cb318', ltk, rand, seq, amf)
    print("Rand: ", bytes.fromhex(rand).hex(","))
    print("Autn: ", bytes.fromhex(autn).hex(","))
    print("CK: ", bytes.fromhex(ck).hex(","))
    print("Res: ", bytes.fromhex(res).hex(","))
    print("Ik: ", bytes.fromhex(ik).hex(","))


functions = dict(
    key_exchange=key_exchange,
    eap_aka=eap_aka,
    decrypt=decrypt,
    encrypt=encrypt,
)

if __name__ == "__main__":
    parser = argparse.ArgumentParser("Encryption experiments")
    subparsers = parser.add_subparsers()
    parser.set_defaults(func=None)
    for name, function in functions.items():
        subparser = subparsers.add_parser(name)
        subparser.set_defaults(func=function)

    args = parser.parse_args()
    print(args)
    if not args.func:
        parser.print_help()
        sys.exit()
    args.func(args)
