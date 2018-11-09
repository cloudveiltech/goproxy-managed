package callbacks

/*
typedef void (*simple_callback)(void* arg);
void fireCallback(void *in, void* arg)
{
	((simple_callback) in)(arg);
}
*/
import "C"
import "unsafe"

func FireCallback(p unsafe.Pointer, arg string) {
	C.fireCallback(unsafe.Pointer(p), unsafe.Pointer(&arg))
}
