package gredis

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	CR             = '\r'
	LF             = '\n'
	CRLF           = "\r\n"
	NoExists       = "-1"
	StatusReply    = '+'
	ErrorReply     = '-'
	IntegerReply   = ':'
	BulkReply      = '$'
	MultiBulkReply = '*'
	Star           = '*'
	Dollar         = '$'
	EmptyString    = ""
)

type Reply struct {
	Err       error
	ReplyType byte
	Bulk      []byte
	SubReply  []Reply
}

func (reply Reply) toString() (string, error) {
	if reply.Err != nil {
		return EmptyString, reply.Err
	}

	if reply.ReplyType == MultiBulkReply {
		return EmptyString, errors.New("return type mismatch, cann't convert to type you want .")
	}

	return string(reply.Bulk), nil
}

func (reply Reply) recurString(prefix string) string {
    if reply.Err != nil {
        return prefix + fmt.Sprintf("%v", reply.Err)
    }
    
    if reply.ReplyType != MultiBulkReply {
        return prefix + string(reply.Bulk)
    }

    if reply.ReplyType == IntegerReply {
        return prefix + " (integer) " + string(reply.Bulk)
    }

    buf := bytes.NewBuffer(nil)
    for i, subreply := range reply.SubReply {
        thisPrefix := fmt.Sprintf("%d) ", i)
        if i == 0 {
            buf.WriteString(prefix)
        } else {
            buf.WriteString(strings.Repeat(" ", len(prefix)))
        }
        buf.WriteString(thisPrefix)
        buf.WriteString(subreply.recurString(strings.Repeat(" ", len(thisPrefix))))
        buf.WriteString("\r\n")
    }
    // delete tail \r\n
    if buf.Len() > 2 {
        return string(buf.Bytes()[0 : buf.Len() - 2])
    }

    return buf.String()
}

func (reply Reply) String() string {
    return reply.recurString(EmptyString)
}

func (reply Reply) Int64() (int64, error) {
	str, err := reply.toString()
	if err != nil {
		return 0, err
	}

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (reply Reply) int32() (int32, error) {
	ui, err := reply.Int64()
	if err != nil {
		return 0, err
	}

	return int32(ui), nil
}

func (reply Reply) Int16() (int16, error) {
	ui, err := reply.Int64()
	if err != nil {
		return 0, err
	}

	return int16(ui), nil
}

func (reply Reply) Int8() (int8, error) {
	ui, err := reply.Uint64()
	if err != nil {
		return 0, err
	}

	return int8(ui), nil
}

func (reply Reply) Uint64() (uint64, error) {
	str, err := reply.toString()
	if err != nil {
		return 0, err
	}

	i, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (reply Reply) Uint32() (uint32, error) {
	ui, err := reply.Uint64()
	if err != nil {
		return 0, err
	}

	return uint32(ui), nil
}

func (reply Reply) Uint16() (uint16, error) {
	ui, err := reply.Uint64()
	if err != nil {
		return 0, err
	}

	return uint16(ui), nil
}

func (reply Reply) Uint8() (uint8, error) {
	ui, err := reply.Uint64()
	if err != nil {
		return 0, err
	}

	return uint8(ui), nil
}

func BuildCmd(args [][]byte) ([]byte, error) {
	argCount := len(args)
	if argCount < 1 {
		return nil, errors.New("argumenr count is too less .")
	}

	buff := bytes.NewBuffer(make([]byte, 0))

	buff.WriteString("*" + strconv.Itoa(argCount) + CRLF)

	for _, arg := range args {
		buff.WriteString("$" + strconv.Itoa(len(args)) + CRLF)
		buff.Write(arg)
		buff.WriteString(CRLF)
	}

	return buff.Bytes(), nil
}

func BuildStringCmd(args []string) ([]byte, error) {
	argCount := len(args)
	if argCount < 1 {
		return nil, errors.New("argumenr count is too less .")
	}

	buff := bytes.NewBuffer(nil)

	buff.WriteString("*" + strconv.Itoa(argCount) + CRLF) // write argument count

	for _, arg := range args {
		buff.WriteString("$" + strconv.Itoa(len(arg)) + CRLF) // write arugment value's byte count
		buff.WriteString(arg + CRLF)                          // write argument value
	}

	return buff.Bytes(), nil
}

func ReadResponse(conn net.Conn) (Reply, error) {
	byteBuff := make([]byte, 1)
	_, err := conn.Read(byteBuff)
	if err != nil {
		return Reply{}, err
	}

	respType := byteBuff[0]

	if respType == StatusReply || respType == ErrorReply || respType == IntegerReply { // status, error, integer response
		line, err := readLine(conn)
		if err != nil {
		    return Reply{}, err
		}
		if respType == ErrorReply {
			return Reply{errors.New(line), ErrorReply, nil, nil}, nil
		}

		return Reply{nil, respType, []byte(line), nil}, nil
	}

	if respType == BulkReply { // bulk response
		bulkData, err := readBulk(conn)
		if err != nil {
		    return Reply{}, err
		}

		return Reply{nil, respType, bulkData, nil}, nil
	}
    
    if respType == MultiBulkReply { // multi bulk response
		return readMultiBulk(conn)
	}

	return Reply{errors.New("unknown response type: " + string([]byte{respType})), respType, nil, nil}, nil
}

func readLine(conn net.Conn) (string, error) {
	byteBuff := make([]byte, 1)
	lineBuff := bytes.NewBuffer(nil)

	for {
		_, err := conn.Read(byteBuff)
		if err != nil {
			return "", err
		}

		b := byteBuff[0]
		if b != CR && b != LF {
			lineBuff.WriteByte(b)
		} else if b == LF {
			break
		}
	}

	return lineBuff.String(), nil
}

func readBulkData(conn net.Conn, bulkLen int) ([]byte, error) {
	bulkBuff := make([]byte, bulkLen+2)

	for i := 0; i < bulkLen+2; {
		cc, err := conn.Read(bulkBuff)
		if err != nil {
			return nil, err
		}
		i += cc
	}
	return bulkBuff[0:bulkLen], nil
}

func readBulkInner(conn net.Conn, hasDoll bool) ([]byte, error) {

	line, err := readLine(conn)
	if err != nil {
		return nil, err
	}
	if hasDoll {
		line = line[1:]
	}

	if strings.EqualFold(line, NoExists) {
		return nil, nil
	}

	bulkLen, err := strconv.Atoi(line)
	if err != nil {
		return nil, err
	}

	bulk, err := readBulkData(conn, bulkLen)
	if err != nil {
		return nil, err
	}

	return bulk, nil
}

func readBulk(conn net.Conn) ([]byte, error) {
	return readBulkInner(conn, false)
}

func readMultiBulk(conn net.Conn) (Reply, error) {
	bulkCountStr, err := readLine(conn)
	if err != nil {
	    return Reply{}, err
	}

	if strings.EqualFold(bulkCountStr, NoExists) {
		return Reply{nil, MultiBulkReply, nil, nil}, nil
	}

	bulkCount, err := strconv.Atoi(bulkCountStr)
	if err != nil {
	    return Reply{}, err
	}

	multiReply := make([]Reply, bulkCount)

	for i := 0; i < bulkCount; i++ {
        subReply, err := ReadResponse(conn)
		if err != nil {
	        return Reply{}, err
		}
        multiReply[i] = subReply
	}

	return Reply{nil, MultiBulkReply, nil, multiReply}, nil
}
