struct MyStruct {
    field: i32,
}

// Method with an immutable reference (cannot modify the struct)
fn immutable_method(s: &MyStruct) {
    println!("{}", s.field);
}

// Method with a mutable reference (can modify the struct)
fn mutable_method(s: &mut MyStruct) {
    s.field = 42;
}

fn main() {
    // normal cases
    let mut my_struct: MyStruct = MyStruct { field: 10 };

    immutable_method(&my_struct);  // Immutable borrow
    println!("{}", my_struct.field);

    mutable_method(&mut my_struct);  // Mutable borrow
    println!("{}", my_struct.field);

    // wrong cases
    let mut mm: MyStruct = MyStruct{ field: 1}; // mm a mutable variable
    mm.field = 2;
    println!("{}", mm.field);

    let m1 = &mut mm; // m1 is a mutable reference of mm
    m1.field = 3;
    println!("{}", m1.field);

    let m2: &MyStruct = m1; // m2 is a imumutable reference of m1, while m1 is a mutable reference of mm, but for safety, this (immutable of mm cannot be passed to m2 from m1)
    //m2.field = 4; // `m2` is a `&` reference, so the data it refers to cannot be written

    let m3: &mut MyStruct = m1;   
    m3.field = 4;
    println!("{}", m3.field); // mm.field is 3, not 4 (because m2 and m3 are two different references

    let mut s1 = Student {name:"Bob".to_string(), age:100};
    println!("{:?}",s1.age);
    s1.set_age(200);
    println!("{}", s1.age);

    // go: 这个case其实是想测试下 pointer value calls non-pointer method (modify fields value)，
    //
    //     1）因为pointer value的methodset是包含了value receivertype and pointer receivertype下的所有方法，它是能够调用到这类，go默认是可以编译通过的
    //     2）但是，且有可能会导致不容易发现的bug，假设是这种写法：
    //     type student struct {
    //         age int
    //     }
    //     假定这里开发不小心写出来一个错误，实际上因为涉及到修改动作，应该使用指针接收器 func(s *student) set_age(a int){ s.age = a}，而不是值接收器
    //     func (s student) set_age(a int){ s.age = a}
    //
    //     假定现在我们现在：
    //     s1 := student{"Bob", 10};
    //     s1.set_age(11);                   // wrong
    //     s2 := &student{"Bob", 20};
    //     s2.set_age(21);                   // wrong
    //
    //     首先可以肯定的是，因为编译都是没有问题的，但是实际执行的时候是不会赋值的，因为值接收器类型，编译器隐式生成的函数实际上是(student) set_age(student, int),
    //     实际上是通过值拷贝copy了一份接收器对象，然后在拷贝的对象上做修改 …… 这个操作在go compiler编译是可以正常通过的，但是实际上容易给开发者埋坑。
    //     ps：尤其是，即便开发者已经将s2定义成指针的情况下，go当前的编译器隐式函数生成操作，也会将其转成对
    //
    //     当前go开发者能做的，也就是：
    //     1. go 目前是可以通过一些ineffective assign来检查，发现一些赋值但是未使用的case，有助于发现此类bug。
    //     2. 时刻谨记，有点提心吊胆，涉及到修改操作，就要将对应的method使用pointer receivertype
    //     3. 或者评估下总是使用pointer receivertype
    //     
    //     ps: 实际上value receivertype对于小对象、没有写操作，是推荐使用的，因为相比于pointer receivertype减少了GC扫描指针的压力。
    //
    // ==============================================================================================================
    //
    // 从这里很容易看出，这是go compiler在这里做了权衡，将压力扔到了开发者这边，这个问题说大不大，说小不小，但是问题一旦发生了而没有事先解决掉，也可能对业务造成大影响。
    // anyway, 我们看看rust是怎么解决这个问题的？
    //
    // 我们假定使用类似go的写法，来写rust代码，rust compiler实际上是会检测出这里存在一个copy的问题，但是我们开发者并没有明确表示这个对象可以被Copy，
    // 那你没有声明可以Copy，如果被Copy了就会发生类似go的问题，容易导致bug。
    // *--: rust而rust直接给你检测出来了这类问题，你不就没有这个风险了吗？
    //
    // 那么rust检测出这个问题，为什么报的是报的这个对象没有实现copy，而不是说，你这里存在一个copy的问题有可能会导致写操作无法reflected出来的问题呢？
    // rustc --explain E0507, 因为编译器不知道你的“原本意图”是什么，但是它可以发现你的“意图+操作”所属于的那个问题大类，更大的类。


    let s2 :&mut Student = &mut s1;
    s2.set_age_v2(210); //
    println!("{}", s2.age);

}

// write a type named student which has two fields: name and age, and implement method set_age(n i32), which sets the age of the student to n.
struct Student{
    name: String,
    age: u8,
}
impl Student {
    fn set_age(&mut self, newAge:u8){
        self.age = newAge;
    }

    //bad1: immutable variable or reference cannot be assigned
    //fn set_age_v2(self, newAge:u8){
    //fn set_age_v2(&self, newAge:u8){


    //bad2: student not implement Copy trait 
    // mut self: so you want to modify it, if so, how do i copy your object? you must do it in copy trait.
    //fn set_age_v2(mut self, newAge:u8) {}

    // good1:
    fn set_age_v2(&mut self, newAge:u8){
        self.age = newAge;
    }

    // good2: implement Copy trait
    // ...
}
